import { FormEvent, useCallback, useEffect, useState } from 'react';
import { api } from '../api/client';
import { ApiError, Category } from '../api/types';
import { useToast } from '../context/ToastContext';

const NAME_MAX = 50;

export default function Categories() {
  const toast = useToast();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newName, setNewName] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const c = await api.listCategories();
      setCategories(Array.isArray(c) ? c : []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to load categories.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    const name = newName.trim();
    if (!name) return;
    if (name.length > NAME_MAX) {
      toast.error(`Name must be ${NAME_MAX} characters or fewer.`);
      return;
    }
    setSubmitting(true);
    try {
      const created = await api.createCategory({ name });
      setCategories((prev) => [...prev, created]);
      setNewName('');
      toast.success('Category added.');
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Failed to create category.');
      // input retained
    } finally {
      setSubmitting(false);
    }
  };

  const startEdit = (c: Category) => {
    setEditingId(c.id);
    setEditingName(c.name);
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditingName('');
  };

  const saveEdit = async (c: Category) => {
    const name = editingName.trim();
    if (!name) {
      toast.error('Name cannot be empty.');
      return;
    }
    if (name.length > NAME_MAX) {
      toast.error(`Name must be ${NAME_MAX} characters or fewer.`);
      return;
    }
    try {
      const updated = await api.updateCategory(c.id, { name });
      setCategories((prev) => prev.map((x) => (x.id === c.id ? { ...x, ...updated } : x)));
      toast.success('Category updated.');
      cancelEdit();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Failed to update category.');
    }
  };

  const handleDelete = async (c: Category) => {
    if (!window.confirm(`Delete "${c.name}"? Tasks using it will become uncategorized.`)) return;
    const prev = categories;
    setCategories((p) => p.filter((x) => x.id !== c.id));
    try {
      await api.deleteCategory(c.id);
      toast.success('Category deleted.');
    } catch (err) {
      setCategories(prev);
      toast.error(err instanceof ApiError ? err.message : 'Failed to delete category.');
    }
  };

  return (
    <div className="categories-page">
      <div className="page-header">
        <h1>Categories</h1>
        <p className="muted">Organize tasks with up to 50 characters per name.</p>
      </div>

      <form className="card inline-form" onSubmit={handleCreate}>
        <input
          type="text"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          maxLength={NAME_MAX}
          placeholder="New category name"
          aria-label="New category name"
          disabled={submitting}
        />
        <button className="btn btn-primary" type="submit" disabled={submitting || !newName.trim()}>
          Add
        </button>
      </form>

      {error && (
        <div className="error-banner" role="alert">
          {error} <button className="btn btn-link" onClick={load}>Retry</button>
        </div>
      )}

      {loading ? (
        <div className="empty">Loading…</div>
      ) : categories.length === 0 ? (
        <div className="empty">No categories yet. Add one above to get started.</div>
      ) : (
        <ul className="list card">
          {categories.map((c) => (
            <li key={c.id} className="list-row">
              {editingId === c.id ? (
                <>
                  <input
                    className="inline-input"
                    type="text"
                    value={editingName}
                    onChange={(e) => setEditingName(e.target.value)}
                    maxLength={NAME_MAX}
                    autoFocus
                  />
                  <div className="list-actions">
                    <button className="btn btn-primary" onClick={() => saveEdit(c)}>Save</button>
                    <button className="btn btn-ghost" onClick={cancelEdit}>Cancel</button>
                  </div>
                </>
              ) : (
                <>
                  <span className="list-name">{c.name}</span>
                  <div className="list-actions">
                    <button className="btn btn-ghost" onClick={() => startEdit(c)}>Rename</button>
                    <button className="btn btn-danger-ghost" onClick={() => handleDelete(c)}>
                      Delete
                    </button>
                  </div>
                </>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
