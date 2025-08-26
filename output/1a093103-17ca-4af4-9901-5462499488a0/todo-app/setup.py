from setuptools import find_packages, setup

setup(
    name="flask_todo_app",
    version="0.1",
    packages=find_packages(),
    include_package_data=True,
    install_requires=[
        "Flask==2.2.2",
        "Flask-SQLAlchemy==3.0.2",
        "Flask-WTF==1.1.1",
        "WTForms==3.0.1",
        "python-dotenv==1.0.0"
    ],
    entry_points={
        "console_scripts": [
            "flask-todo=run:main"
        ]
    }
)