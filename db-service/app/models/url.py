from sqlalchemy.orm import DeclarativeBase
from sqlalchemy import String, Text, Integer, DateTime, func, Column

class Base(DeclarativeBase):
    pass

class URL(Base):
    __tablename__ = "urls"
    __table_args__ = {"schema": "tiny_url"}

    id = Column(Integer, primary_key=True)
    original_url = Column(Text, nullable=False)
    short_code = Column(String(15), unique=True, nullable=False)
    ip_addr = Column(String(15), nullable=False)
    exp_time = Column(Integer, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
