<script>
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores.js';
  import { api } from '$lib/api.js';

  let email = '';
  let password = '';
  let confirmPassword = '';
  let error = '';
  let loading = false;
  let apiKey = '';
  let done = false;
  let copied = false;

  async function setup() {
    if (!email || !password || !confirmPassword) {
      error = 'Please fill in all fields';
      return;
    }
    if (password !== confirmPassword) {
      error = 'Passwords do not match';
      return;
    }
    if (password.length < 8) {
      error = 'Password must be at least 8 characters';
      return;
    }
    loading = true;
    error = '';
    try {
      const res = await api.post('/api/auth/setup', { email, password });
      apiKey = res.apiKey;
      localStorage.setItem('token', res.token);
      authStore.set({ token: res.token, authenticated: true });
      done = true;
    } catch (e) {
      error = e.message;
    }
    loading = false;
  }

  async function copyKey() {
    await navigator.clipboard.writeText(apiKey);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }
</script>

<div class="min-h-screen bg-[#0d0d14] flex items-center justify-center px-4">
  <div class="w-full max-w-sm">
    <div class="text-center mb-8">
      <div class="w-11 h-11 bg-blue-600 rounded-xl flex items-center justify-center text-lg font-bold text-white mx-auto mb-3">P</div>
      <h1 class="text-xl font-semibold text-white">PicoWatch</h1>
      <p class="text-slate-500 text-sm mt-1">Create your admin account</p>
    </div>

    {#if done}
      <div class="bg-[#12121e] border border-[#1a1a2e] rounded-xl p-6">
        <div class="text-center mb-5">
          <div class="w-10 h-10 bg-emerald-500/15 rounded-full flex items-center justify-center mx-auto mb-3">
            <svg class="w-5 h-5 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 class="text-base font-semibold text-white">Account created</h2>
          <p class="text-slate-500 text-xs mt-1">Save your API key — it won't be shown again.</p>
        </div>

        <div class="bg-[#0d0d14] border border-[#1a1a2e] rounded-lg p-3 mb-4">
          <div class="text-xs text-slate-500 mb-1.5">API Key</div>
          <div class="font-mono text-xs text-blue-300 break-all leading-relaxed">{apiKey}</div>
        </div>

        <div class="flex gap-2 mb-4">
          <button on:click={copyKey} class="flex-1 bg-[#1a1a2e] hover:bg-[#232340] text-slate-300 text-xs py-2 rounded-lg transition-colors border border-[#232340]">
            {copied ? 'Copied!' : 'Copy Key'}
          </button>
          <button on:click={() => goto('/')} class="flex-1 bg-blue-600 hover:bg-blue-500 text-white text-xs py-2 rounded-lg transition-colors">
            Open Dashboard
          </button>
        </div>

        <div class="bg-[#0d0d14] border border-[#1a1a2e] rounded-lg p-3 text-xs">
          <div class="text-slate-500 mb-1.5">Usage</div>
          <code class="text-slate-400">Authorization: Bearer <span class="text-blue-300">{apiKey.slice(0, 16)}...</span></code>
        </div>
      </div>
    {:else}
      <div class="bg-[#12121e] border border-[#1a1a2e] rounded-xl p-6">
        {#if error}
          <div class="bg-red-500/10 border border-red-500/30 rounded-lg px-4 py-2.5 text-red-400 text-sm mb-4">
            {error}
          </div>
        {/if}

        <form on:submit|preventDefault={setup} class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-1.5">Email</label>
            <input type="email" bind:value={email} placeholder="admin@example.com" required class="input w-full" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-1.5">Password</label>
            <input type="password" bind:value={password} placeholder="Min. 8 characters" required class="input w-full" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-1.5">Confirm Password</label>
            <input type="password" bind:value={confirmPassword} placeholder="••••••••" required class="input w-full" />
          </div>
          <button
            type="submit"
            disabled={loading}
            class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium py-2.5 rounded-lg transition-colors text-sm"
          >
            {loading ? 'Creating account...' : 'Create Account'}
          </button>
        </form>
      </div>
    {/if}
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
