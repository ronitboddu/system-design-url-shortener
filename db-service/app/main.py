from core.config import Settings
from core.database import DatabaseManager
from repositories.url_repository import URLRepository
from fastapi import FastAPI

settings = Settings()
db = DatabaseManager(settings)
url_repository = URLRepository(db.SessionLocal)

app = FastAPI()
app.state.url_repository = url_repository

db.create_schema()
db.create_tables()
db.clear_tables()