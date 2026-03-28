from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    app_name: str = "db-service"
    app_env: str = "development"

    db_host: str = "localhost" 
    db_port: int = 5432
    db_name: str = "postgres"
    db_user: str = "ronitboddu"
    db_password: str = ""
    db_schema: str = "tiny_url"

    def database_url(self) -> str:
        return (
            f"postgresql+psycopg://{self.db_user}"
            f"@{self.db_host}:{self.db_port}/{self.db_name}"
        )
    