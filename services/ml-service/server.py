import grpc
import ml_service_pb2
from concurrent.futures import ThreadPoolExecutor

import classifier
import ner
import pricing
from ml_service_pb2_grpc import MLServiceServicer, add_MLServiceServicer_to_server


class MLServicer(MLServiceServicer):
    def AnalyzeListing(self, request, context):
        entities = ner.extract_entities(request.title, request.description)
        violation, _ = classifier.detect_violation(request.description)

        suggested, lower, upper = pricing.predict_price(
            title=request.title,
            desc=request.description,
            condition=request.condition,
            shipping=0.0,
            brand=entities["brand"],
            category=request.category_id,
        )

        return ml_service_pb2.ListingAnalysis(
            brand=entities["brand"],
            product=entities["product"],
            size=entities["size"],
            policy_violation=violation,
            suggested_price=suggested,
            price_lower_bound=lower,
            price_upper_bound=upper,
        )


def serve():
    server = grpc.server(ThreadPoolExecutor(max_workers=10))
    add_MLServiceServicer_to_server(MLServicer(), server)
    server.add_insecure_port("[::]:50051")
    server.start()
    print("MLService server started on port 50051")
    server.wait_for_termination()


if __name__ == "__main__":
    serve()