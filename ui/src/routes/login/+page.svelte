<script>
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores.js';
  import { api } from '$lib/api.js';

  let email = '';
  let password = '';
  let error = '';
  let loading = false;

  async function login() {
    if (!email || !password) {
      error = 'Please fill in all fields';
      return;
    }
    loading = true;
    error = '';
    try {
      const { token } = await api.post('/api/auth/login', { email, password });
      localStorage.setItem('token', token);
      authStore.set({ token, authenticated: true });
      await goto('/');
    } catch (e) {
      error = e.message;
    }
    loading = false;
  }
</script>

<div class="min-h-screen bg-[#0d0d14] flex items-center justify-center px-4">
  <div class="w-full max-w-sm">
    <div class="text-center mb-8">
      <h1 class="text-xl font-semibold text-white">PicoWatch</h1>
      <p class="text-slate-500 text-sm mt-1">Sign in to your account</p>
    </div>

    <div class="bg-[#12121e] border border-[#1a1a2e] rounded-xl p-6">
      {#if error}
        <div class="bg-red-500/10 border border-red-500/30 rounded-lg px-4 py-2.5 text-red-400 text-sm mb-4">
          {error}
        </div>
      {/if}

      <form on:submit|preventDefault={login} class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-1.5">Email</label>
          <input
            type="email"
            bind:value={email}
            placeholder="admin@example.com"
            required
            class="input w-full"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-1.5">Password</label>
          <input
            type="password"
            bind:value={password}
            placeholder="••••••••"
            required
            class="input w-full"
          />
        </div>
        <button
          type="submit"
          disabled={loading}
          class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium py-2.5 rounded-lg transition-colors text-sm"
        >
          {loading ? 'Signing in...' : 'Sign in'}
        </button>
      </form>
    </div>
  </div>
</div>

<style>
  .input {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    padding: 0.5rem 0.75rem;
    color: var(--text);
    font-size: 0.875rem;
    transition: border-color 0.15s;
    outline: none;
    display: block;
  }
  .input::placeholder { color: var(--text-placeholder); }
  .input:focus { border-color: #6366f1; }
</style>
