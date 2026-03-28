from pydantic import BaseModel

class CreateURLRequest(BaseModel):
    original_url: str
    short_code: str
    exp_time: int

class URLResponse(BaseModel):
    original_url: str
    short_code: str
    exp_time: int