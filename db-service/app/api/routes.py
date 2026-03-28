from fastapi import APIRouter, Request, HTTPException
from schemas.url import CreateURLRequest, URLResponse

router = APIRouter()

@router.get("/urls/{short_code}")
def get_url(short_code: str, request: Request):
    repo = request.app.state.url_repository
    record = repo.get_record(short_code)
    if not record:
        raise HTTPException(status_code=404, detail="not found")
    return URLResponse(
        original_url=record.original_url,
        short_code=record.short_code,
        exp_time=record.exp_time,
    )

@router.post("/urls")
def create_url(payload: CreateURLRequest, request: Request):
    repo = request.app.state.url_repository
    record = repo.put_record(
        original_url=payload.original_url,
        short_code=payload.short_code,
        exp_time=payload.exp_time,
    )
    return URLResponse(
        original_url=record.original_url,
        short_code=record.short_code,
        exp_time=record.exp_time,
    )
