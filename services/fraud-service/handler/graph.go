package handler

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"

    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
    "github.com/vexyruu/LIP/shared/apitypes"
)

var errUserNotFound = errors.New("user not found")

// GetGraph returns the graph for a given user.
func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	nodes, edges, err := h.queryUserGraph(r.Context(), userID)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			http.Error(w, "user not found in graph", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get graph", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apitypes.GraphResponse{
		UserID: userID,
		Nodes: nodes,
		Edges: edges,
	})


}

// queryUserGraph queries the graph for a given user.
func (h *Handler) queryUserGraph(ctx context.Context, userID string) ([]apitypes.GraphNode, []apitypes.GraphEdge, error) {
	session := h.Neo4j.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Check if the user exists
	check, err := session.Run(ctx, `
		MATCH (u:User {user_id: $userID}) RETURN u LIMIT 1`,
		map[string]any{"userID": userID},
	)
	if err != nil {
		return nil, nil, err
	}
	if !check.Next(ctx) {
		return nil, nil, errUserNotFound
	}

	// Query the graph for the user and its neighbors
	result, err := session.Run(ctx, `
		MATCH (u:User {user_id: $userID})
		OPTIONAL MATCH (u)-[:USED_DEVICE|SHARED_IP|USED_PAYMENT*1..2]-(n)
		WITH u, collect(DISTINCT n) AS neighbors
		WITH [u] + neighbors AS rawNodes
		UNWIND rawNodes AS node
		WITH collect(DISTINCT node) AS allNodes
		UNWIND allNodes AS n1
		UNWIND allNodes AS n2
		OPTIONAL MATCH (n1)-[r:USED_DEVICE|SHARED_IP|USED_PAYMENT]-(n2)
		WHERE elementId(n1) < elementId(n2)
		RETURN allNodes, collect(DISTINCT r) AS rels
	`, map[string]any{"userID": userID})
	if err != nil {
		return nil, nil, err
	}
	// Handle no result
	if !result.Next(ctx) {
		return []apitypes.GraphNode{}, []apitypes.GraphEdge{}, nil
	}

	// Get the record
	record := result.Record()
	rawNodes, _ := record.Get("allNodes")
	rawRels, _ := record.Get("rels")

	// Initialize the maps
	idByElement := map[string]string{}
	nodesByID := map[string]apitypes.GraphNode{}

	// Handle nodes
	if nodeList, ok := rawNodes.([]any); ok {
		for _, item := range nodeList {
			node, ok := item.(dbtype.Node)
			if !ok {
				continue
			}
			gn, ok := graphNodeFromNeo4jNode(node)
			if !ok {
				continue
			}
			nodesByID[gn.ID] = gn
			idByElement[node.ElementId] = gn.ID
		}
	}

	// Handle edges
	edgeSeen := map[string]struct{}{}
	var edges []apitypes.GraphEdge

	if relList, ok := rawRels.([]any); ok {
		for _, item := range relList {
			rel, ok := item.(dbtype.Relationship)
			if !ok {
				continue
			}
			edge, ok := graphEdgeFromNeo4jRelationship(rel, idByElement)
			if !ok {
				continue
			}
			key := edge.Source + "|" + edge.Type + "|" + edge.Target
			if _, dup := edgeSeen[key]; dup {
				continue
			}
			edgeSeen[key] = struct{}{}
			edges = append(edges, edge)
		}
	}
	// Return the nodes and edges
	nodes := make([]apitypes.GraphNode, 0, len(nodesByID))
	for _, n := range nodesByID {
		nodes = append(nodes, n)
	}

	return nodes, edges, nil
}

// graphNodeFromNeo4jNode converts a Neo4j node to a GraphNode.
func graphNodeFromNeo4jNode(node dbtype.Node) (apitypes.GraphNode, bool) {
	props := node.Props
	labels := node.Labels
	if len(labels) == 0 {
		return apitypes.GraphNode{}, false
	}

	switch labels[0] {
		case "User":
			id, _ := props["user_id"].(string)
			gn := apitypes.GraphNode{ID: id, Type: "User", Label: id}
			if banned, ok := props["banned"].(bool); ok {
				gn.Banned = &banned
			}
			if rs, ok := props["risk_score"].(float64); ok {
				gn.RiskScore = &rs
			}
			return gn, id != ""
		case "Device":
			id, _ := props["device_fingerprint"].(string)
			return apitypes.GraphNode{ID: id, Type: "Device", Label: id}, id != ""
		case "IPAddress":
			id, _ := props["ip"].(string)
			return apitypes.GraphNode{ID: id, Type: "IPAddress", Label: id}, id != ""
		case "PaymentMethod":
			id, _ := props["payment_id"].(string)
			return apitypes.GraphNode{ID: id, Type: "PaymentMethod", Label: id}, id != ""
		default:
			return apitypes.GraphNode{}, false
	}
}

// graphEdgeFromNeo4jRelationship converts a Neo4j relationship to a GraphEdge.
func graphEdgeFromNeo4jRelationship(rel dbtype.Relationship, idByElement map[string]string) (apitypes.GraphEdge,bool) {
	source := idByElement[rel.StartElementId]
	target := idByElement[rel.EndElementId]
	
	if source == "" || target == "" {
		return apitypes.GraphEdge{}, false
	}
	return apitypes.GraphEdge{
		Source: source, 
		Target: target, 
		Type: rel.Type,
	}, true	
}