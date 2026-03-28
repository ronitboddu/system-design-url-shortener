from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    app_name: str = "db-service"
    app_env: str = "development"

    db_host: str = "localhost" 
    db_port: int = 5432
    db_name: str = "postgres"
    db_user: str = "ronitboddu"
    db_password: str = ""
    db_schema: str = "tiny_url"

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    @property
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