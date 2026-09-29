import os
from typing import List

SUPPORTED_MODELS: List[str] = [
    "mistral-small-latest",
    "mistral-medium-latest",
    "mistral-large-latest",
    "open-mistral-7b",
    "open-mixtral-8x7b",
    "open-mixtral-8x22b",
    "codestral-latest",
]

class WorkerConfig:
    def __init__(self):
        self.worker_id = os.getenv("WORKER_ID", "inference-worker-1")
        self.host = os.getenv("GRPC_HOST", "0.0.0.0")
        self.port = int(os.getenv("GRPC_PORT", "50051"))
        self.mistral_api_key = os.getenv("MISTRAL_API_KEY", "")
        self.max_workers = int(os.getenv("GRPC_MAX_WORKERS", "10"))
        self.supported_models = SUPPORTED_MODELS

config = WorkerConfig()
