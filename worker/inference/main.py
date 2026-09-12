import sys
import asyncio
from pathlib import Path
import logging
import signal

# Ensure project root is in python path
root_dir = str(Path(__file__).resolve().parent.parent.parent)
if root_dir not in sys.path:
    sys.path.insert(0, root_dir)

import grpc
from worker.inference.inference import inference_pb2_grpc
from worker.inference.app.servicer import InferenceServicer
from worker.inference.models.config import config

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s - %(message)s",
)
logger = logging.getLogger("inference_worker")


async def serve():
    """
    Async gRPC server using grpc.aio.

    All concurrent inference streams are handled cooperatively on the asyncio
    event loop — no OS thread per request. Concurrency is bounded by Mistral
    API throughput and the Go-side admission controller, not thread pool size.
    """
    server = grpc.aio.server()
    inference_pb2_grpc.add_InferenceServiceServicer_to_server(
        InferenceServicer(), server
    )
    bind_addr = f"{config.host}:{config.port}"
    server.add_insecure_port(bind_addr)

    logger.info(f"Starting Inference Worker '{config.worker_id}' on {bind_addr} (grpc.aio)")
    await server.start()
    logger.info(f"Supported models: {config.supported_models}")

    async def shutdown():
        logger.info("Graceful shutdown: draining in-flight requests (5s grace)...")
        await server.stop(grace=5)
        logger.info("Worker stopped.")

    loop = asyncio.get_event_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, lambda: asyncio.ensure_future(shutdown()))

    await server.wait_for_termination()


if __name__ == "__main__":
    asyncio.run(serve())
