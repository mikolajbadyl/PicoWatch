<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import '../app.css';
  import { authStore } from '$lib/stores.js';
  import { api } from '$lib/api.js';

  let loading = true;
  let sidebarOpen = false;
  let theme = 'dark';

  function applyTheme(t) {
    document.documentElement.classList.toggle('light', t === 'light');
  }

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark';
    localStorage.setItem('theme', theme);
    applyTheme(theme);
  }

  const publicRoutes = ['/login', '/setup'];

  onMount(async () => {
    const saved = localStorage.getItem('theme');
    if (saved) {
      theme = saved;
    } else if (window.matchMedia('(prefers-color-scheme: light)').matches) {
      theme = 'light';
    }
    applyTheme(theme);

    window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', (e) => {
      if (!localStorage.getItem('theme')) {
        theme = e.matches ? 'light' : 'dark';
        applyTheme(theme);
      }
    });

    try {
      const { hasAdmin } = await api.request('GET', '/api/auth/status');

      if (!hasAdmin) {
        if ($page.url.pathname !== '/setup') {
          await goto('/setup');
        }
        loading = false;
        return;
      }

      if ($page.url.pathname === '/setup') {
        await goto('/login');
        loading = false;
        return;
      }

      const token = localStorage.getItem('token');
      if (!token) {
        if ($page.url.pathname !== '/login') {
          await goto('/login');
        }
        loading = false;
        return;
      }

      authStore.set({ token, authenticated: true });
    } catch (e) {
      console.error('Auth check failed:', e);
    }
    loading = false;
  });


  $: if ($page.url.pathname) {
    sidebarOpen = false;
  }

  function logout() {
    localStorage.removeItem('token');
    authStore.set({ token: null, authenticated: false });
    goto('/login');
  }
</script>

{#if loading}
  <div class="flex items-center justify-center h-screen bg-[#0d0d14]">
    <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
  </div>
{:else if publicRoutes.includes($page.url.pathname)}
  <slot />
{:else if $authStore.authenticated}
  <div class="flex h-screen overflow-hidden bg-[#0d0d14]">

    {#if sidebarOpen}
      <div
        class="fixed inset-0 z-30 bg-black/60 md:hidden"
        on:click={() => (sidebarOpen = false)}
        role="presentation"
      ></div>
    {/if}

    <aside class="sidebar" class:open={sidebarOpen}>
      <div class="p-4 border-b border-[#1a1a2e] flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <span class="font-semibold text-white text-sm">PicoWatch</span>
        </div>
        <button
          class="md:hidden text-slate-400 hover:text-slate-200 p-1"
          on:click={() => (sidebarOpen = false)}
          aria-label="Close menu"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <nav class="flex-1 p-2 pt-3">
        <a href="/" class="nav-item" class:active={$page.url.pathname === '/'}>
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          Home
        </a>
        <a href="/stats" class="nav-item" class:active={$page.url.pathname === '/stats'}>
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          Stats
        </a>
        <a href="/logs" class="nav-item" class:active={$page.url.pathname === '/logs'}>
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          Logs
        </a>
        <a href="/settings" class="nav-item" class:active={$page.url.pathname === '/settings'}>
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Settings
        </a>
      </nav>

      <div class="p-2 border-t border-[#1a1a2e]">
        <button
          on:click={toggleTheme}
          class="w-full text-left text-sm text-slate-400 hover:text-slate-200 transition-colors flex items-center gap-2.5 px-3 py-2 rounded-lg hover:bg-[#1a1a2e] mb-0.5"
        >
          {#if theme === 'dark'}
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707M17.657 17.657l-.707-.707M6.343 6.343l-.707-.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            Light mode
          {:else}
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
            Dark mode
          {/if}
        </button>
        <button
          on:click={logout}
          class="w-full text-left text-sm text-slate-400 hover:text-slate-200 transition-colors flex items-center gap-2.5 px-3 py-2 rounded-lg hover:bg-[#1a1a2e]"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          Logout
        </button>
      </div>
    </aside>

    <div class="flex flex-col flex-1 min-w-0 overflow-hidden">
      <header class="md:hidden flex items-center gap-3 px-4 py-3 bg-[#12121e] border-b border-[#1a1a2e] shrink-0">
        <button
          on:click={() => (sidebarOpen = true)}
          class="text-slate-400 hover:text-slate-200 p-1"
          aria-label="Open menu"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <div class="flex items-center gap-2">
          <span class="font-semibold text-white text-sm">PicoWatch</span>
        </div>
      </header>

      <main class="flex-1 overflow-auto">
        <slot />
      </main>
    </div>
  </div>
{:else}
  <div class="flex items-center justify-center h-screen bg-[#0d0d14]">
    <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
  </div>
{/if}

<style>
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 40;
    width: 13rem;
    background: var(--surface);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    transform: translateX(-100%);
    transition: transform 0.25s ease;
  }

  .sidebar.open {
    transform: translateX(0);
  }

  @media (min-width: 768px) {
    .sidebar {
      position: relative;
      transform: translateX(0);
      flex-shrink: 0;
    }
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.5rem 0.75rem;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    color: var(--text-muted);
    transition: all 0.15s;
    width: 100%;
    margin-bottom: 0.125rem;
    text-decoration: none;
  }
  .nav-item:hover {
    color: var(--text);
    background-color: var(--raised);
  }
  .nav-item.active {
    color: var(--nav-active-text);
    background-color: var(--nav-active-bg);
    font-weight: 500;
  }
</style>
