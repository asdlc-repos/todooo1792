import { FormEvent, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { api } from '../api/client';
import { ApiError, Category } from '../api/types';
import { useToast } from '../context/ToastContext';
import { fromDateTimeLocal, toDateTimeLocal } from '../utils/dateUtils';

const TITLE_MAX = 200;
const DESC_MAX = 2000;

export default function TaskEdit() {
  const { id } = useParams<{ id: string }>();
  const isNew = !id;
  const navigate = useNavigate();
  const toast = useToast();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [categoryId, setCategoryId] = useState<string>('');
  const [dueDate, setDueDate] = useState<string>('');
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(!isNew);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let ignore = false;
    (async () => {
      try {
        const cats = await api.listCategories();
        if (!ignore) setCategories(Array.isArray(cats) ? cats : []);
      } catch {
        /* non-fatal; form still usable without categories */
      }
    })();
    return () => {
      ignore = true;
    };
  }, []);

  useEffect(() => {
    if (isNew) return;
    let ignore = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const t = await api.getTask(id!);
        if (ignore) return;
        setTitle(t.title || '');
        setDescription(t.description || '');
        setCategoryId(t.categoryId || '');
        setDueDate(toDateTimeLocal(t.dueDate));
      } catch (err) {
        if (!ignore) setError(err instanceof ApiError ? err.message : 'Failed to load task.');
      } finally {
        if (!ignore) setLoading(false);
      }
    })();
    return () => {
      ignore = true;
    };
  }, [id, isNew]);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      setError('Title is required.');
      return;
    }
    if (trimmedTitle.length > TITLE_MAX) {
      setError(`Title must be ${TITLE_MAX} characters or fewer.`);
      return;
    }
    if (description.length > DESC_MAX) {
      setError(`Description must be ${DESC_MAX} characters or fewer.`);
      return;
    }
    setSubmitting(true);
    const payload = {
      title: trimmedTitle,
      description: description.trim() || undefined,
      categoryId: categoryId || null,
      dueDate: fromDateTimeLocal(dueDate),
    };
    try {
      if (isNew) {
        await api.createTask(payload);
        toast.success('Task created.');
      } else {
        await api.updateTask(id!, payload);
        toast.success('Task updated.');
      }
      navigate('/');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to save task.');
      // Inputs retained by design — do not reset state on error.
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <div className="empty">Loading task…</div>;

  return (
    <div className="form-page">
      <div className="page-header">
        <h1>{isNew ? 'New task' : 'Edit task'}</h1>
      </div>
      <form className="card form" onSubmit={onSubmit} noValidate>
        <label className="field">
          <span>Title</span>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            maxLength={TITLE_MAX}
            required
            disabled={submitting}
            autoFocus
          />
          <small className="hint">{title.length}/{TITLE_MAX}</small>
        </label>
        <label className="field">
          <span>Description</span>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            maxLength={DESC_MAX}
            rows={5}
            disabled={submitting}
          />
          <small className="hint">{description.length}/{DESC_MAX}</small>
        </label>
        <div className="field-row">
          <label className="field">
            <span>Category</span>
            <select
              value={categoryId}
              onChange={(e) => setCategoryId(e.target.value)}
              disabled={submitting}
            >
              <option value="">No category</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>Due date &amp; time</span>
            <input
              type="datetime-local"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              disabled={submitting}
            />
          </label>
        </div>
        {error && <div className="form-error" role="alert">{error}</div>}
        <div className="form-actions">
          <button type="button" className="btn btn-ghost" onClick={() => navigate(-1)} disabled={submitting}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? 'Saving…' : isNew ? 'Create task' : 'Save changes'}
          </button>
        </div>
      </form>
    </div>
  );
}
