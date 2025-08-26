from flask_wtf import FlaskForm
from wtforms import StringField, TextAreaField, BooleanField, SubmitField
from wtforms.validators import DataRequired, Length

class TodoForm(FlaskForm):
    title = StringField("Title", validators=[DataRequired(), Length(min=1, max=140)])
    description = TextAreaField("Description")
    completed = BooleanField("Completed")
    submit = SubmitField("Submit")