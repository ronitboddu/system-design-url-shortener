from sqlalchemy import create_engine, inspect, text
from sqlalchemy.orm import sessionmaker
from sqlalchemy.schema import CreateSchema

from app.core.config import Settings
from app.models.url import Base

class DatabaseManager:
    def __init__(self, settings: Settings):
        self.settings = settings
        self.engine = create_engine(settings.database_url(), echo=False)
        self.SessionLocal = sessionmaker(bind=self.engine)

    def create_schema(self) -> None:
        schema_name = self.settings.db_schema

        with self.engine.begin() as conn:
            if not inspect(conn).has_schema(schema_name):
                conn.execute(CreateSchema(schema_name))
        print("tiny_url schema created successfully!")
    
    def create_tables(self) -> None:
        Base.metadata.create_all(bind=self.engine)
        print("Table 'urls' created successfully!")

    def clear_tables(self) -> None:
        with self.engine.begin() as conn:
            conn.execute(text("TRUNCATE TABLE tiny_url.urls RESTART IDENTITY"))
            print("Table is empty!")
