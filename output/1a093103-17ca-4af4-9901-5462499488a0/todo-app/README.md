This is a simple Todo application built with Flask.

Features:
- Blueprint-based architecture for modular design
- SQLAlchemy for database operations
- Form validation using Flask-WTF
- Environment-based configuration (development, production, etc.)
- Custom error handling for 404 and 500 errors

Requirements:
- Python 3.6+
- See requirements.txt for project dependencies

How to Run the Application:
1. Create and activate a virtual environment:
   - On Linux/Mac:
       python3 -m venv venv
       source venv/bin/activate
   - On Windows:
       python -m venv venv
       venv\Scripts\activate

2. Install dependencies:
       pip install -r requirements.txt

3. Setup environment variables:
   - For development, you can set the environment variable FLASK_ENV to "development". For example:
       export FLASK_ENV=development   (Linux/Mac)
       set FLASK_ENV=development      (Windows)

4. Initialize the database:
   - Open a Python shell with your virtual environment active:
         python
     Then run:
         >>> from app import db, create_app
         >>> app = create_app()
         >>> app.app_context().push()
         >>> db.create_all()
         >>> exit()

5. Run the application:
       flask run

Alternatively, you can run the app using:
       python run.py

Directory Structure:
.
├── README.md
├── requirements.txt
├── setup.py
├── run.py
├── config.py
└── app
    ├── __init__.py
    ├── models.py
    ├── todo
    │   ├── __init__.py
    │   ├── forms.py
    │   └── routes.py
    └── templates
        ├── base.html
        ├── index.html
        ├── add_todo.html
        ├── 404.html
        └── 500.html

Happy coding!