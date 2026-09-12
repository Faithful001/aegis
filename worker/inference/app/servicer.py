import logging
from worker.inference.inference import inference_pb2, inference_pb2_grpc
from worker.inference.chains.inference_chain import inference_chain
from worker.inference.models.config import config

logger = logging.getLogger("inference_servicer")


class InferenceServicer(inference_pb2_grpc.InferenceServiceServicer):
    """
    Async gRPC servicer for InferenceService.

    Runs inside grpc.aio — each Generate() call is a native coroutine.
    All concurrent streams are multiplexed on the same asyncio event loop,
    so concurrency is limited by Mistral API throughput, not OS thread count.
    """

    async def Generate(self, request, context):
        logger.info(
            "Received Generate request",
            extra={
                "req_id": request.request_id,
                "job_id": request.job_id,
                "model": request.model,
                "stream": request.stream,
                "org_id": request.organization_id,
                "project_id": request.project_id,
            },
        )
        async for response in inference_chain.execute(request):
            yield response

    async def HealthCheck(self, request, context):
        return inference_pb2.HealthCheckResponse(
            status="SERVING",
            supported_models=config.supported_models,
        )
