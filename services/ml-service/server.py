import grpc, ml_service_pb2
from concurrent.futures import ThreadPoolExecutor
from ml_service_pb2_grpc import MLServiceServicer, add_MLServiceServicer_to_server

# MLServiceServicer implementation
class MLServicer(MLServiceServicer):
    def AnalyzeListing(self, request, context):
        return ml_service_pb2.ListingAnalysis(
            brand="", product="", size="",
            policy_violation=False,
            suggested_price=0.0,
            price_lower_bound=0.0,
            price_upper_bound=0.0
        )

# Serve the MLService
def serve():
    server = grpc.server(ThreadPoolExecutor(max_workers=10))
    add_MLServiceServicer_to_server(MLServicer(), server)
    server.add_insecure_port("[::]:50051")
    server.start()
    print("MLService server started on port 50051")
    server.wait_for_termination()
    
if __name__ == "__main__":
    serve()
