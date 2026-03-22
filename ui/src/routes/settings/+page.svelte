<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';

  let tab = 'users';

  let pricingStatus = null;
  let pricingSyncing = false;
  let pricingMsg = '';

  let retentionDays = '';
  let retentionSaving = false;
  let retentionMsg = '';

  // Users
  let users = [];
  let usersLoading = true;
  let newEmail = '';
  let newPassword = '';
  let addingUser = false;
  let addUserError = '';
  let addUserSuccess = '';

  // Password change modal
  let changePwdUserId = null;
  let changePwdEmail = '';
  let changePwdValue = '';
  let changePwdLoading = false;
  let changePwdError = '';

  let apiKeys = [];
  let keysLoading = true;
  let newKeyName = '';
  let addingKey = false;
  let newKeyValue = '';
  let addKeyError = '';

  let me = null;

  onMount(async () => {
    await Promise.all([loadUsers(), loadKeys(), loadMe(), loadPricingStatus(), loadRetention()]);
  });

  async function loadPricingStatus() {
    try { pricingStatus = await api.get('/api/pricing/status'); } catch {}
  }

  async function loadRetention() {
    try {
      const res = await api.get('/api/retention');
      retentionDays = res.retention_days || '';
    } catch {}
  }

  async function saveRetention() {
    retentionSaving = true;
    retentionMsg = '';
    try {
      await api.post('/api/retention', { retention_days: retentionDays });
      retentionMsg = 'Saved';
      setTimeout(() => retentionMsg = '', 3000);
    } catch (e) {
      retentionMsg = e.message;
    }
    retentionSaving = false;
  }

  async function syncPricing() {
    pricingSyncing = true;
    pricingMsg = '';
    try {
      const res = await api.post('/api/pricing/sync', {});
      pricingMsg = `Synced ${res.model_count} models`;
      await loadPricingStatus();
    } catch (e) {
      pricingMsg = e.message;
    }
    pricingSyncing = false;
  }

  async function loadMe() {
    try { me = await api.get('/api/me'); } catch {}
  }

  async function loadUsers() {
    usersLoading = true;
    try { users = await api.get('/api/users'); } catch {}
    usersLoading = false;
  }

  async function loadKeys() {
    keysLoading = true;
    try { apiKeys = await api.get('/api/apikeys'); } catch {}
    keysLoading = false;
  }

  async function addUser() {
    addUserError = '';
    addUserSuccess = '';
    if (!newEmail || !newPassword) { addUserError = 'Fill in all fields'; return; }
    if (newPassword.length < 8) { addUserError = 'Password min. 8 characters'; return; }
    addingUser = true;
    try {
      await api.post('/api/users', { email: newEmail, password: newPassword });
      newEmail = '';
      newPassword = '';
      addUserSuccess = 'User created';
      setTimeout(() => addUserSuccess = '', 3000);
      await loadUsers();
    } catch (e) {
      addUserError = e.message;
    }
    addingUser = false;
  }

  async function deleteUser(id) {
    if (!confirm('Delete this user?')) return;
    try {
      await api.del(`/api/users/${id}`);
      await loadUsers();
    } catch (e) {
      alert(e.message);
    }
  }

  function openChangePwd(user) {
    changePwdUserId = user.id;
    changePwdEmail = user.email;
    changePwdValue = '';
    changePwdError = '';
  }

  async function changePassword() {
    changePwdError = '';
    if (changePwdValue.length < 8) { changePwdError = 'Min. 8 characters'; return; }
    changePwdLoading = true;
    try {
      await api.request('PATCH', `/api/users/${changePwdUserId}/password`, { password: changePwdValue });
      changePwdUserId = null;
    } catch (e) {
      changePwdError = e.message;
    }
    changePwdLoading = false;
  }

  async function addKey() {
    addKeyError = '';
    newKeyValue = '';
    addingKey = true;
    try {
      const res = await api.post('/api/apikeys', { name: newKeyName || 'API Key' });
      newKeyValue = res.key;
      newKeyName = '';
      await loadKeys();
    } catch (e) {
      addKeyError = e.message;
    }
    addingKey = false;
  }

  async function deleteKey(id) {
    if (!confirm('Delete this API key?')) return;
    try {
      await api.del(`/api/apikeys/${id}`);
      await loadKeys();
    } catch (e) {
      alert(e.message);
    }
  }

  let copiedKey = false;
  async function copyKey(val) {
    await navigator.clipboard.writeText(val);
    copiedKey = true;
    setTimeout(() => copiedKey = false, 2000);
  }

  function fmtDate(str) {
    if (!str) return '—';
    return new Date(str).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
</script>

<!-- Change password modal -->
{#if changePwdUserId !== null}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-black/60" on:click={() => changePwdUserId = null} role="presentation"></div>
    <div class="relative bg-[#12121e] border border-[#1a1a2e] rounded-xl p-6 w-full max-w-sm">
      <h3 class="text-white font-semibold mb-1">Change password</h3>
      <p class="text-slate-500 text-xs mb-4">{changePwdEmail}</p>

      {#if changePwdError}
        <div class="bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2 text-red-400 text-xs mb-3">{changePwdError}</div>
      {/if}

      <input
        type="password"
        bind:value={changePwdValue}
        placeholder="New password (min. 8 chars)"
        class="input w-full mb-4"
        on:keydown={(e) => e.key === 'Enter' && changePassword()}
      />

      <div class="flex gap-2 justify-end">
        <button on:click={() => changePwdUserId = null} class="btn-ghost">Cancel</button>
        <button on:click={changePassword} disabled={changePwdLoading} class="btn-primary">
          {changePwdLoading ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  </div>
{/if}

<div class="p-4 sm:p-6 max-w-3xl mx-auto">
  <div class="mb-6">
    <h1 class="text-xl font-semibold text-white">Settings</h1>
    <p class="text-slate-500 text-sm mt-0.5">Manage users and API keys</p>
  </div>

  <div class="flex flex-wrap gap-1 mb-6 bg-[#12121e] border border-[#1a1a2e] rounded-xl p-1 w-fit">
    <button class="tab-btn" class:active={tab === 'users'} on:click={() => tab = 'users'}>
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
      Users
    </button>
    <button class="tab-btn" class:active={tab === 'apikeys'} on:click={() => tab = 'apikeys'}>
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" /></svg>
      API Keys
    </button>
    <button class="tab-btn" class:active={tab === 'pricing'} on:click={() => tab = 'pricing'}>
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      Pricing
    </button>
    <button class="tab-btn" class:active={tab === 'retention'} on:click={() => tab = 'retention'}>
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      Retention
    </button>
    <button class="tab-btn" class:active={tab === 'snippets'} on:click={() => tab = 'snippets'}>
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>
      Snippets
    </button>
  </div>

  <!-- USERS TAB -->
  {#if tab === 'users'}
    <!-- Add user form -->
    <div class="panel mb-5">
      <div class="section-title">Add user</div>

      {#if addUserError}
        <div class="bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2 text-red-400 text-xs mb-3">{addUserError}</div>
      {/if}
      {#if addUserSuccess}
        <div class="bg-emerald-500/10 border border-emerald-500/30 rounded-lg px-3 py-2 text-emerald-400 text-xs mb-3">{addUserSuccess}</div>
      {/if}

      <div class="flex flex-col sm:flex-row gap-2">
        <input
          type="email"
          bind:value={newEmail}
          placeholder="Email"
          class="input flex-1"
        />
        <input
          type="password"
          bind:value={newPassword}
          placeholder="Password (min. 8)"
          class="input flex-1"
          on:keydown={(e) => e.key === 'Enter' && addUser()}
        />
        <button on:click={addUser} disabled={addingUser} class="btn-primary shrink-0">
          {addingUser ? 'Adding...' : 'Add'}
        </button>
      </div>
    </div>

    <!-- Users list -->
    <div class="panel">
      <div class="section-title">Users</div>

      {#if usersLoading}
        <div class="flex justify-center py-8">
          <div class="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
        </div>
      {:else if users.length === 0}
        <div class="text-slate-500 text-sm text-center py-6">No users</div>
      {:else}
        <div class="space-y-2">
          {#each users as user}
            <div class="flex items-center justify-between gap-3 bg-[#0d0d14] rounded-lg px-3 py-2.5">
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-7 h-7 rounded-full bg-blue-600/20 flex items-center justify-center text-blue-300 text-xs font-semibold shrink-0">
                  {user.email[0].toUpperCase()}
                </div>
                <div class="min-w-0">
                  <div class="text-sm text-slate-200 truncate">{user.email}</div>
                  <div class="text-xs text-slate-500">Joined {fmtDate(user.created_at)}</div>
                </div>
                {#if me && me.id === user.id}
                  <span class="text-xs text-blue-400 bg-blue-500/10 px-1.5 py-0.5 rounded shrink-0">you</span>
                {/if}
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                <button
                  on:click={() => openChangePwd(user)}
                  class="icon-btn"
                  title="Change password"
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                {#if !me || me.id !== user.id}
                  <button
                    on:click={() => deleteUser(user.id)}
                    class="icon-btn danger"
                    title="Delete user"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- API KEYS TAB -->
  {#if tab === 'apikeys'}
    <!-- New key shown once -->
    {#if newKeyValue}
      <div class="bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-4 mb-5">
        <div class="text-xs text-emerald-400 font-medium mb-2">New API key — copy it now, won't be shown again</div>
        <div class="bg-[#0d0d14] rounded-lg px-3 py-2.5 font-mono text-xs text-blue-300 break-all mb-3">{newKeyValue}</div>
        <button on:click={() => copyKey(newKeyValue)} class="btn-ghost text-xs">
          {copiedKey ? 'Copied!' : 'Copy'}
        </button>
      </div>
    {/if}

    <!-- Add key form -->
    <div class="panel mb-5">
      <div class="section-title">Create API key</div>
      {#if addKeyError}
        <div class="bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2 text-red-400 text-xs mb-3">{addKeyError}</div>
      {/if}
      <div class="flex gap-2">
        <input
          type="text"
          bind:value={newKeyName}
          placeholder="Key name (e.g. Production)"
          class="input flex-1"
          on:keydown={(e) => e.key === 'Enter' && addKey()}
        />
        <button on:click={addKey} disabled={addingKey} class="btn-primary shrink-0">
          {addingKey ? 'Creating...' : 'Create'}
        </button>
      </div>
    </div>

    <!-- Keys list -->
    <div class="panel">
      <div class="section-title">Active keys</div>

      {#if keysLoading}
        <div class="flex justify-center py-8">
          <div class="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
        </div>
      {:else if apiKeys.length === 0}
        <div class="text-slate-500 text-sm text-center py-6">No API keys</div>
      {:else}
        <div class="space-y-2">
          {#each apiKeys as key}
            <div class="flex items-center justify-between gap-3 bg-[#0d0d14] rounded-lg px-3 py-2.5">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 mb-0.5">
                  <span class="text-sm text-slate-200">{key.name}</span>
                  <span class="text-xs text-slate-500 font-mono">{key.key_preview}</span>
                </div>
                <div class="text-xs text-slate-500">
                  {key.user_email} · Created {fmtDate(key.created_at)}
                  {#if key.last_used_at}
                    · Last used {fmtDate(key.last_used_at)}
                  {:else}
                    · Never used
                  {/if}
                </div>
              </div>
              <button
                on:click={() => deleteKey(key.id)}
                class="icon-btn danger shrink-0"
                title="Delete key"
              >
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- PRICING TAB -->
  {#if tab === 'pricing'}
    <div class="panel">
      <div class="section-title">Model pricing</div>
      <p class="text-slate-500 text-xs mb-4">
        Prices fetched automatically from <span class="text-blue-400">OpenRouter</span> — refreshed every 24h.
        Cost is calculated automatically if not provided in the request.
      </p>

      <div class="flex items-center justify-between bg-[#0d0d14] rounded-lg px-4 py-3 mb-4">
        <div>
          {#if pricingStatus}
            <div class="text-sm text-slate-200 font-medium">{pricingStatus.model_count} models</div>
            <div class="text-xs text-slate-500 mt-0.5">
              Last sync: {pricingStatus.updated_at ? new Date(pricingStatus.updated_at).toLocaleString() : 'never'}
            </div>
          {:else}
            <div class="text-sm text-slate-500">Loading...</div>
          {/if}
        </div>
        <button on:click={syncPricing} disabled={pricingSyncing} class="btn-primary">
          {#if pricingSyncing}
            <span class="flex items-center gap-2">
              <span class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin inline-block"></span>
              Syncing...
            </span>
          {:else}
            Sync now
          {/if}
        </button>
      </div>

      {#if pricingMsg}
        <div class="bg-blue-500/10 border border-blue-500/30 rounded-lg px-3 py-2 text-blue-300 text-xs">{pricingMsg}</div>
      {/if}

      <div class="mt-4 bg-[#0d0d14] rounded-lg p-3 text-xs text-slate-400 space-y-1">
        <div class="font-medium text-slate-300 mb-2">How auto-cost works:</div>
        <div>• Send <span class="font-mono text-blue-300">cost: 0</span> or omit the field — cost is calculated automatically</div>
        <div>• Model matching: exact → suffix (<span class="font-mono text-blue-300">gpt-4o</span> → <span class="font-mono text-blue-300">openai/gpt-4o</span>) → substring</div>
        <div>• To override — just send <span class="font-mono text-blue-300">cost</span> in the request</div>
      </div>
    </div>
  {/if}

  <!-- RETENTION TAB -->
  {#if tab === 'retention'}
    <div class="panel">
      <div class="section-title">Log retention</div>
      <p class="text-slate-500 text-xs mb-4">
        Automatically delete logs older than the specified number of days. Leave empty to keep logs indefinitely.
      </p>

      <div class="flex items-center gap-3 mb-3">
        <input
          type="number"
          min="1"
          bind:value={retentionDays}
          placeholder="e.g. 90"
          class="input w-32"
        />
        <span class="text-slate-400 text-sm">days</span>
        <button on:click={saveRetention} disabled={retentionSaving} class="btn-primary">
          {retentionSaving ? 'Saving...' : 'Save'}
        </button>
      </div>

      {#if retentionMsg}
        <div class="bg-emerald-500/10 border border-emerald-500/30 rounded-lg px-3 py-2 text-emerald-400 text-xs">{retentionMsg}</div>
      {/if}

      <div class="mt-4 bg-[#0d0d14] rounded-lg p-3 text-xs text-slate-400">
        Cleanup runs once every 24h. Setting to empty disables automatic deletion.
      </div>
    </div>
  {/if}

  <!-- SNIPPETS TAB -->
  {#if tab === 'snippets'}
    <div class="space-y-4">
      <div class="panel">
        <div class="section-title">curl</div>
        <pre class="snippet">{`curl -X POST https://your-domain/api/logs \\
  -H "Authorization: Bearer pk_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4o",
    "input_tokens": 1200,
    "output_tokens": 340,
    "status": "success",
    "metadata": { "user": "alice", "session": "abc123" }
  }'`}</pre>
      </div>

      <div class="panel">
        <div class="section-title">Python</div>
        <pre class="snippet">{`import requests

requests.post(
    "https://your-domain/api/logs",
    headers={"Authorization": "Bearer pk_your_api_key"},
    json={
        "model": "gpt-4o",
        "input_tokens": 1200,
        "output_tokens": 340,
        "status": "success",
        "metadata": {"user": "alice"},
    },
)`}</pre>
      </div>

      <div class="panel">
        <div class="section-title">JavaScript / TypeScript</div>
        <pre class="snippet">{`await fetch("https://your-domain/api/logs", {
  method: "POST",
  headers: {
    "Authorization": "Bearer pk_your_api_key",
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    model: "gpt-4o",
    input_tokens: 1200,
    output_tokens: 340,
    status: "success",
    metadata: { user: "alice" },
  }),
});`}</pre>
      </div>

      <div class="panel">
        <div class="section-title">Request fields</div>
        <div class="overflow-x-auto">
          <table class="w-full text-xs">
            <thead>
              <tr class="border-b border-[#1a1a2e] text-left">
                <th class="pb-2 text-slate-500 font-medium pr-4">Field</th>
                <th class="pb-2 text-slate-500 font-medium pr-4">Type</th>
                <th class="pb-2 text-slate-500 font-medium">Description</th>
              </tr>
            </thead>
            <tbody class="text-slate-400">
              <tr class="border-b border-[#1a1a2e]/40">
                <td class="py-2 pr-4 font-mono text-blue-300">model</td>
                <td class="py-2 pr-4">string</td>
                <td class="py-2">Required. Model name e.g. <span class="font-mono">gpt-4o</span></td>
              </tr>
              <tr class="border-b border-[#1a1a2e]/40">
                <td class="py-2 pr-4 font-mono text-blue-300">input_tokens</td>
                <td class="py-2 pr-4">int</td>
                <td class="py-2">Number of prompt tokens</td>
              </tr>
              <tr class="border-b border-[#1a1a2e]/40">
                <td class="py-2 pr-4 font-mono text-blue-300">output_tokens</td>
                <td class="py-2 pr-4">int</td>
                <td class="py-2">Number of completion tokens</td>
              </tr>
              <tr class="border-b border-[#1a1a2e]/40">
                <td class="py-2 pr-4 font-mono text-blue-300">cost</td>
                <td class="py-2 pr-4">float</td>
                <td class="py-2">Cost in USD. Auto-calculated from OpenRouter if omitted</td>
              </tr>
              <tr class="border-b border-[#1a1a2e]/40">
                <td class="py-2 pr-4 font-mono text-blue-300">status</td>
                <td class="py-2 pr-4">string</td>
                <td class="py-2"><span class="font-mono">success</span> or <span class="font-mono">error</span> (default: success)</td>
              </tr>
              <tr class="border-b border-[#1a1a2e]/40">
                <td class="py-2 pr-4 font-mono text-blue-300">error</td>
                <td class="py-2 pr-4">string</td>
                <td class="py-2">Error message if status is error</td>
              </tr>
              <tr>
                <td class="py-2 pr-4 font-mono text-blue-300">metadata</td>
                <td class="py-2 pr-4">object</td>
                <td class="py-2">Any custom JSON — user, session, tags, etc.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .panel {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    padding: 1.25rem;
  }
  .section-title {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.875rem;
  }
  .input {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    padding: 0.5rem 0.75rem;
    color: var(--text);
    font-size: 0.875rem;
    transition: border-color 0.15s;
    outline: none;
  }
  .input::placeholder { color: var(--text-placeholder); }
  .input:focus { border-color: #6366f1; }
  .btn-primary {
    background: #4f46e5;
    color: white;
    font-size: 0.8125rem;
    padding: 0.5rem 0.875rem;
    border-radius: 0.5rem;
    transition: background 0.15s;
    border: none;
    cursor: pointer;
  }
  .btn-primary:hover:not(:disabled) { background: #6366f1; }
  .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-ghost {
    background: var(--raised);
    color: var(--text-muted);
    font-size: 0.8125rem;
    padding: 0.5rem 0.875rem;
    border-radius: 0.5rem;
    transition: all 0.15s;
    border: none;
    cursor: pointer;
  }
  .btn-ghost:hover { background: var(--raised2); color: var(--text); }
  .tab-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.875rem;
    border-radius: 0.625rem;
    font-size: 0.8125rem;
    color: var(--text-dim);
    background: transparent;
    border: none;
    cursor: pointer;
    transition: all 0.15s;
    white-space: nowrap;
  }
  .tab-btn:hover { color: var(--text-muted); }
  .tab-btn.active {
    background: var(--raised);
    color: var(--text);
  }
  .icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
    height: 1.75rem;
    border-radius: 0.375rem;
    background: var(--raised);
    color: var(--text-dim);
    border: none;
    cursor: pointer;
    transition: all 0.15s;
  }
  .icon-btn:hover { background: var(--raised2); color: var(--text-muted); }
  .icon-btn.danger:hover { background: rgba(239,68,68,0.15); color: #f87171; }
  .snippet {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    padding: 0.875rem;
    font-size: 0.75rem;
    color: var(--text-muted);
    overflow-x: auto;
    line-height: 1.6;
    white-space: pre;
  }
</style>
