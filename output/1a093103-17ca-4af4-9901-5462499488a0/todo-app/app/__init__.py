from flask import Flask, render_template
from flask_sqlalchemy import SQLAlchemy
from config import config_by_name

db = SQLAlchemy()

def create_app(config_name="development"):
    app = Flask(__name__)
    app.config.from_object(config_by_name[config_name])
    
    # Initialize Flask extensions
    db.init_app(app)
    
    # Register Blueprints
    from app.todo import bp as todo_bp
    app.register_blueprint(todo_bp, url_prefix="/todos")
    
    # Error Handlers
    register_error_handlers(app)
    
    return app

def register_error_handlers(app):
    @app.errorhandler(404)
    def not_found_error(error):
        return render_template("404.html"), 404

    @app.errorhandler(500)
    def internal_error(error):
        # Rollback in case of database errors
        db.session.rollback()
        return render_template("500.html"), 500