from sqlalchemy.orm import DeclarativeBase
from sqlalchemy import String, Text, Integer, DateTime, func, Column, Index

class Base(DeclarativeBase):
    pass

class URL(Base):
    __tablename__ = "urls"
    __table_args__ = (
        Index(
            "urls_original_url_ip_addr_idx",
            "original_url",
            "ip_addr",
            unique=True,
        ),
        {"schema": "tiny_url"},
    )

    id = Column(Integer, primary_key=True)
    original_url = Column(Text, nullable=False)
    short_code = Column(String(15), unique=True, nullable=False)
    ip_addr = Column(String(15), nullable=False)
    exp_time = Column(Integer, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
