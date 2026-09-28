import { useEffect, useState } from "react";

const API = "/api/todos";

export default function App() {
  const [todos, setTodos] = useState([]);
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");

  async function load() {
    const res = await fetch(API);
    if (!res.ok) throw new Error("не удалось загрузить список");
    setTodos(await res.json());
  }

  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, []);

  async function add(e) {
    e.preventDefault();
    const t = title.trim();
    if (!t) return;
    setError("");
    const res = await fetch(API, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title: t }),
    });
    if (!res.ok) {
      setError("не удалось добавить задачу");
      return;
    }
    setTitle("");
    await load();
  }

  async function toggle(todo) {
    const res = await fetch(`${API}/${todo.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ completed: !todo.completed }),
    });
    if (res.ok) await load();
  }

  async function remove(id) {
    const res = await fetch(`${API}/${id}`, { method: "DELETE" });
    if (res.ok) await load();
  }

  return (
    <main className="page">
      <h1>Todo list</h1>
      <form onSubmit={add} className="row">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Новая задача"
        />
        <button type="submit">Добавить</button>
      </form>
      {error && <p className="error">{error}</p>}
      <ul>
        {todos.map((todo) => (
          <li key={todo.id} className={todo.completed ? "done" : ""}>
            <label>
              <input
                type="checkbox"
                checked={todo.completed}
                onChange={() => toggle(todo)}
              />
              {todo.title}
            </label>
            <button type="button" onClick={() => remove(todo.id)}>
              Удалить
            </button>
          </li>
        ))}
      </ul>
    </main>
  );
}
