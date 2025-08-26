import os

basedir = os.path.abspath(os.path.dirname(__file__))

class Config:
    SECRET_KEY = os.environ.get("SECRET_KEY", "this_should_be_changed")
    SQLALCHEMY_DATABASE_URI = os.environ.get("DATABASE_URL") or \
        'sqlite:///' + os.path.join(basedir, 'app.db')
    SQLALCHEMY_TRACK_MODIFICATIONS = False

class DevelopmentConfig(Config):
    DEBUG = True

class ProductionConfig(Config):
    DEBUG = False

config_by_name = dict(
    development=DevelopmentConfig,
    production=ProductionConfig
)