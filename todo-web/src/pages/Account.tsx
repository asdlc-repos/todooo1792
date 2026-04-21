import { FormEvent, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { ApiError } from '../api/types';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';

export default function Account() {
  const { user, updateUser, logout } = useAuth();
  const toast = useToast();
  const navigate = useNavigate();

  const [email, setEmail] = useState(user?.email || '');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    const trimmedEmail = email.trim();
    if (!trimmedEmail) {
      setError('Email cannot be empty.');
      return;
    }
    const wantsPasswordChange = !!newPassword || !!confirm;
    if (wantsPasswordChange) {
      if (newPassword.length < 8) {
        setError('New password must be at least 8 characters.');
        return;
      }
      if (newPassword !== confirm) {
        setError('New passwords do not match.');
        return;
      }
      if (!currentPassword) {
        setError('Enter your current password to change it.');
        return;
      }
    }
    setSubmitting(true);
    const payload: { email?: string; password?: string; currentPassword?: string } = {};
    if (trimmedEmail !== user?.email) payload.email = trimmedEmail;
    if (wantsPasswordChange) {
      payload.password = newPassword;
      payload.currentPassword = currentPassword;
    }
    if (Object.keys(payload).length === 0) {
      toast.notify('No changes to save.');
      setSubmitting(false);
      return;
    }
    try {
      const updated = await api.updateAccount(payload);
      if (updated && typeof updated === 'object' && 'email' in updated) {
        updateUser(updated);
      } else if (payload.email) {
        updateUser({ id: user?.id || '', email: payload.email });
      }
      setCurrentPassword('');
      setNewPassword('');
      setConfirm('');
      toast.success('Account updated.');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update account.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <div className="form-page">
      <div className="page-header">
        <h1>Account</h1>
        <p className="muted">Update your email or change your password.</p>
      </div>
      <form className="card form" onSubmit={onSubmit} noValidate>
        <label className="field">
          <span>Email</span>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
            disabled={submitting}
          />
        </label>
        <fieldset className="fieldset">
          <legend>Change password</legend>
          <label className="field">
            <span>Current password</span>
            <input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              autoComplete="current-password"
              disabled={submitting}
            />
          </label>
          <label className="field">
            <span>New password</span>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              autoComplete="new-password"
              minLength={8}
              disabled={submitting}
            />
          </label>
          <label className="field">
            <span>Confirm new password</span>
            <input
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              autoComplete="new-password"
              disabled={submitting}
            />
          </label>
        </fieldset>
        {error && <div className="form-error" role="alert">{error}</div>}
        <div className="form-actions">
          <button type="button" className="btn btn-danger-ghost" onClick={handleLogout} disabled={submitting}>
            Log out
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? 'Saving…' : 'Save changes'}
          </button>
        </div>
      </form>
    </div>
  );
}
