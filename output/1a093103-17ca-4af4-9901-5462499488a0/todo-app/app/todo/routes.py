from flask import render_template, redirect, url_for, flash, request
from app import db
from app.models import Todo
from app.todo import bp
from app.todo.forms import TodoForm

@bp.route("/", methods=["GET"])
def index():
    todos = Todo.query.all()
    return render_template("index.html", todos=todos)

@bp.route("/add", methods=["GET", "POST"])
def add_todo():
    form = TodoForm()
    if form.validate_on_submit():
        todo = Todo(
            title=form.title.data,
            description=form.description.data,
            completed=form.completed.data
        )
        db.session.add(todo)
        db.session.commit()
        flash("Todo added successfully.", "success")
        return redirect(url_for("todo.index"))
    return render_template("add_todo.html", form=form)

@bp.route("/<int:todo_id>/edit", methods=["GET", "POST"])
def edit_todo(todo_id):
    todo = Todo.query.get_or_404(todo_id)
    form = TodoForm(obj=todo)
    if form.validate_on_submit():
        todo.title = form.title.data
        todo.description = form.description.data
        todo.completed = form.completed.data
        db.session.commit()
        flash("Todo updated successfully.", "success")
        return redirect(url_for("todo.index"))
    return render_template("add_todo.html", form=form, todo=todo)

@bp.route("/<int:todo_id>/delete", methods=["POST"])
def delete_todo(todo_id):
    todo = Todo.query.get_or_404(todo_id)
    db.session.delete(todo)
    db.session.commit()
    flash("Todo deleted successfully.", "success")
    return redirect(url_for("todo.index"))