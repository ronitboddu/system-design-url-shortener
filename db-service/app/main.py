from app.core.config import Settings
from app.core.database import DatabaseManager
from app.repositories.url_repository import URLRepository
from app.util.snowflake import SnowflakeGenerator
from fastapi import FastAPI
from app.api.routes import router

settings = Settings()
db = DatabaseManager(settings)
generator = SnowflakeGenerator(node_id=settings.snowflake_node_id)
url_repository = URLRepository(db.SessionLocal, generator)

app = FastAPI()
app.state.url_repository = url_repository
app.include_router(router)

db.create_schema()
db.create_tables()
# db.clear_tables()