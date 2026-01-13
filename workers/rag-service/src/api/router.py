from fastapi import APIRouter

from api.v1 import documents
from api.v1 import retrieval

api_router_v1 = APIRouter(
    prefix="/v1",
    tags=["v1"],
    include_in_schema=True,
)

api_router_v1.include_router(
    documents.router,
    prefix="/documents",
    tags=["documents"],
    include_in_schema=True,
)

api_router_v1.include_router(
    retrieval.router,
    prefix="/qa",
    tags=["retrieval"],
    include_in_schema=True,
)
