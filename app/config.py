"""Application configuration."""
import os
from dotenv import load_dotenv

load_dotenv()


class Config:
    """Base configuration."""
    SECRET_KEY = os.environ.get('SECRET_KEY', 'dev-secret-key-change-in-production')
    SECURITY_PASSWORD_SALT = os.environ.get('SECURITY_PASSWORD_SALT', 'dev-salt-change-in-production')
    
    # Database
    SQLALCHEMY_DATABASE_URI = os.environ.get('DATABASE_URL', 'mysql+pymysql://vehicle_auc:vehicle_auc@localhost:3306/vehicle_auc')
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    SQLALCHEMY_ENGINE_OPTIONS = {
        'pool_pre_ping': True,
        'pool_recycle': 300,
    }
    
    # Redis
    REDIS_URL = os.environ.get('REDIS_URL', 'redis://localhost:6379/0')
    
    # Flask-Security-Too
    SECURITY_REGISTERABLE = True
    SECURITY_CONFIRMABLE = True
    SECURITY_RECOVERABLE = True
    SECURITY_TRACKABLE = True
    SECURITY_CHANGEABLE = True
    SECURITY_SEND_REGISTER_EMAIL = True
    SECURITY_EMAIL_SENDER = os.environ.get('SENDGRID_FROM_EMAIL', 'noreply@localhost')
    
    # Session
    SESSION_TYPE = 'redis'
    PERMANENT_SESSION_LIFETIME = 86400  # 24 hours
    
    # CSRF
    WTF_CSRF_ENABLED = True
    
    # File uploads
    MAX_CONTENT_LENGTH = 16 * 1024 * 1024  # 16MB max
    
    # AWS S3
    AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
    AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
    AWS_S3_BUCKET = os.environ.get('AWS_S3_BUCKET')
    AWS_REGION = os.environ.get('AWS_REGION', 'us-east-1')
    
    # External APIs
    AUTHORIZE_NET_API_LOGIN_ID = os.environ.get('AUTHORIZE_NET_API_LOGIN_ID')
    AUTHORIZE_NET_TRANSACTION_KEY = os.environ.get('AUTHORIZE_NET_TRANSACTION_KEY')
    AUTHORIZE_NET_SANDBOX = os.environ.get('AUTHORIZE_NET_SANDBOX', 'true').lower() == 'true'
    
    PERSONA_API_KEY = os.environ.get('PERSONA_API_KEY')
    PERSONA_TEMPLATE_ID = os.environ.get('PERSONA_TEMPLATE_ID')
    
    CLEARVIN_API_KEY = os.environ.get('CLEARVIN_API_KEY')
    
    SUPER_DISPATCH_API_KEY = os.environ.get('SUPER_DISPATCH_API_KEY')
    
    DLR_DMV_API_KEY = os.environ.get('DLR_DMV_API_KEY')
    
    SENDGRID_API_KEY = os.environ.get('SENDGRID_API_KEY')
    SENDGRID_FROM_EMAIL = os.environ.get('SENDGRID_FROM_EMAIL', 'noreply@localhost')
    
    OPENAI_API_KEY = os.environ.get('OPENAI_API_KEY')


class DevelopmentConfig(Config):
    """Development configuration."""
    DEBUG = True
    SQLALCHEMY_ECHO = True
    SECURITY_SEND_REGISTER_EMAIL = False  # Disable email in dev


class TestingConfig(Config):
    """Testing configuration."""
    TESTING = True
    DEBUG = True
    SQLALCHEMY_DATABASE_URI = os.environ.get(
        'TEST_DATABASE_URL', 
        'mysql+pymysql://vehicle_auc_test:vehicle_auc_test@localhost:3307/vehicle_auc_test'
    )
    REDIS_URL = os.environ.get('TEST_REDIS_URL', 'redis://localhost:6380/0')
    WTF_CSRF_ENABLED = False
    SECURITY_SEND_REGISTER_EMAIL = False
    SERVER_NAME = 'localhost:5000'


class ProductionConfig(Config):
    """Production configuration."""
    DEBUG = False
    SQLALCHEMY_ECHO = False
    
    # Ensure these are set in production
    @property
    def SECRET_KEY(self):
        key = os.environ.get('SECRET_KEY')
        if not key:
            raise ValueError('SECRET_KEY must be set in production')
        return key


config = {
    'development': DevelopmentConfig,
    'testing': TestingConfig,
    'production': ProductionConfig,
    'default': DevelopmentConfig,
}
