from fastapi import APIRouter, Request, HTTPException
from app.schemas.url import CreateURLRequest, URLResponse
import time

router = APIRouter()

@router.get("/urls/{short_code}")
def get_url(short_code: str, request: Request):
    start_time = time.perf_counter()
    repo = request.app.state.url_repository
    record = repo.get_record(short_code)
    if not record:
        raise HTTPException(status_code=404, detail="not found")
    end_time = time.perf_counter()
    print(f"get URL Execution time: {end_time - start_time:.4f} seconds")
    return URLResponse(
        original_url=record["original_url"],
        short_code=record["short_code"],
        exp_time=record["exp_time"],
    )

@router.post("/urls")
def create_url(payload: CreateURLRequest, request: Request):
    start_time = time.perf_counter()
    repo = request.app.state.url_repository
    record = repo.put_record(
        original_url=payload.original_url,
        ip_addr=payload.ip_addr,
        exp_time=payload.exp_time,
    )

    end_time = time.perf_counter()
    print(f"create URL Execution time: {end_time - start_time:.4f} seconds") 
    return URLResponse(
        original_url=record["original_url"],
        short_code=record["short_code"],
        exp_time=record["exp_time"],
    )
