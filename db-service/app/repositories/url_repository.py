from app.models.url import URL
from app.util.base62 import encode_base62

class URLRepository:
    def __init__(self, session_factory, generator):
        self.session_factory = session_factory
        self.generator = generator

    def put_record(self, original_url, ip_addr, exp_time):
        with self.session_factory() as session:
            with session.begin():
                try:
                    existing = (
                        session.query(URL)
                        .filter(URL.original_url == original_url,
                                URL.ip_addr == ip_addr)
                        .first()
                    )

                    if existing:
                        # print(f"{existing.short_code} already exists in the DB.\n")
                        return {
                        "original_url": existing.original_url,
                        "short_code": existing.short_code,
                        "exp_time": existing.exp_time,
                    }
                    
                    
                    id = self.generator.next_id()
                    short_code = encode_base62(id)

                    record = URL(
                        original_url=original_url,
                        short_code=short_code,
                        ip_addr = ip_addr,
                        exp_time=exp_time,
                    )
                    session.add(record)
                    return  {
                        "original_url": record.original_url,
                        "short_code": record.short_code,
                        "exp_time": record.exp_time,
                    }
                except:
                    session.rollback()
                    raise
        
    def get_record(self, short_code):
        with self.session_factory() as session:
            with session.begin():
                try:
                    record = session.query(URL).filter(URL.short_code == short_code).first()
                    
                    if not record:
                        return None
                    
                    return {
                        "original_url": record.original_url,
                        "short_code": record.short_code,
                        "exp_time": record.exp_time,
                    }
                except:
                    session.rollback()
                    raise
