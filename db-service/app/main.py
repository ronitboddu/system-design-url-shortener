from sqlalchemy import create_engine

from core.config import Settings
from core.database import DatabaseManager

settings = Settings()
db = DatabaseManager(settings)

db.create_schema()
db.create_tables()