from models.url import URL

class URLRepository:
    def __init__(self, session_factory):
        self.session_factory = session_factory

    def put_record(self, original_url, short_code, exp_time):
        with self.session_factory() as session:
            record = URL(
                original_url=original_url,
                short_code=short_code,
                exp_time=exp_time,
            )
            session.add(record)
            session.commit()
            session.refresh(record)
            return record
        
    def get_record(self, short_code):
        with self.session_factory() as session:
            return session.query(URL).filter(URL.short_code == short_code).first()
