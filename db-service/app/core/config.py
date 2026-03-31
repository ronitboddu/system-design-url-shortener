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
    snowflake_node_id: int = 1

    def database_url(self) -> str:
        if self.db_password:
            return (
                f"postgresql+psycopg://{self.db_user}:{self.db_password}"
                f"@{self.db_host}:{self.db_port}/{self.db_name}"
            )
        return (
            f"postgresql+psycopg://{self.db_user}"
            f"@{self.db_host}:{self.db_port}/{self.db_name}"
        )