<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';

  let logs = [];
  let total = 0;
  let page = 1;
  let limit = 50;
  let modelFilter = '';
  let dateRange = 'all';
  let loading = true;
  let error = null;
  let sortCol = 'created_at';
  let sortDir = 'desc';
  let expandedId = null;

  const dateRanges = [
    { label: 'All', value: 'all' },
    { label: 'Today', value: 'today' },
    { label: '7d', value: '7d' },
    { label: '30d', value: '30d' },
  ];

  function getDateParams() {
    const now = new Date();
    const pad = (n) => String(n).padStart(2, '0');
    const fmt = (d) => `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}`;
    if (dateRange === 'today') return { from: fmt(now), to: fmt(now) };
    if (dateRange === '7d') {
      const d = new Date(now); d.setDate(d.getDate() - 6);
      return { from: fmt(d), to: fmt(now) };
    }
    if (dateRange === '30d') {
      const d = new Date(now); d.setDate(d.getDate() - 29);
      return { from: fmt(d), to: fmt(now) };
    }
    return {};
  }

  async function load() {
    loading = true;
    error = null;
    try {
      const params = new URLSearchParams({ page, limit, sort: sortCol, dir: sortDir });
      if (modelFilter) params.set('model', modelFilter);
      const dp = getDateParams();
      if (dp.from) params.set('from', dp.from);
      if (dp.to) params.set('to', dp.to);
      const res = await api.get('/api/logs?' + params);
      logs = res.logs ?? [];
      total = res.total ?? 0;
    } catch (e) {
      error = e.message;
    }
    loading = false;
  }

  onMount(load);

  function applyFilter() { page = 1; load(); }
  function clearFilter() { modelFilter = ''; page = 1; load(); }
  function setRange(r) { dateRange = r; page = 1; load(); }
  function prevPage() { if (page > 1) { page--; load(); } }
  function nextPage() { if (page < totalPages) { page++; load(); } }

  function toggleExpand(id) {
    expandedId = expandedId === id ? null : id;
  }

  function fmtDate(str) {
    return new Date(str).toLocaleString('en-US', {
      month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit', second: '2-digit',
    });
  }

  function fmtCost(n) {
    if (!n) return '—';
    return '$' + n.toFixed(6);
  }

  function parseMeta(raw) {
    try {
      const obj = JSON.parse(raw);
      if (!obj || typeof obj !== 'object' || Array.isArray(obj) || Object.keys(obj).length === 0) return null;
      return obj;
    } catch { return null; }
  }

  function metaVal(v) {
    if (Array.isArray(v)) return v.join(', ');
    if (typeof v === 'object' && v !== null) return JSON.stringify(v);
    return String(v);
  }

  function setSort(col) {
    if (sortCol === col) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortCol = col;
      sortDir = 'desc';
    }
    page = 1;
    load();
  }

  async function exportCSV() {
    const params = new URLSearchParams();
    if (modelFilter) params.set('model', modelFilter);
    const dp = getDateParams();
    if (dp.from) params.set('from', dp.from);
    if (dp.to) params.set('to', dp.to);
    const token = localStorage.getItem('token');
    const res = await fetch('/api/logs/export?' + params, {
      headers: { Authorization: 'Bearer ' + token },
    });
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  $: totalPages = Math.ceil(total / limit);
</script>

<div class="p-4 sm:p-6">
  <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3 mb-5">
    <div>
      <h1 class="text-xl font-semibold text-white">Logs</h1>
      <p class="text-slate-500 text-sm mt-0.5">{total.toLocaleString()} requests</p>
    </div>

    <div class="flex flex-col gap-2">
      <div class="flex items-center gap-2">
        <input
          type="text"
          bind:value={modelFilter}
          placeholder="Filter by model..."
          on:keydown={(e) => e.key === 'Enter' && applyFilter()}
          class="bg-[#12121e] border border-[#1a1a2e] focus:border-blue-500 rounded-lg px-3 py-1.5 text-sm text-slate-200 placeholder-slate-600 focus:outline-none transition-colors flex-1 sm:w-48 sm:flex-none min-w-0"
        />
        <button on:click={applyFilter} class="btn-primary shrink-0">Filter</button>
        {#if modelFilter}
          <button on:click={clearFilter} class="btn-ghost shrink-0">Clear</button>
        {/if}
        <button on:click={exportCSV} class="btn-ghost shrink-0">Export CSV</button>
      </div>
      <div class="flex gap-1">
        {#each dateRanges as r}
          <button
            on:click={() => setRange(r.value)}
            class="range-btn"
            class:active={dateRange === r.value}
          >{r.label}</button>
        {/each}
      </div>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-24">
      <div class="w-7 h-7 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if error}
    <div class="bg-red-500/10 border border-red-500/30 rounded-xl p-4 text-red-400 text-sm">{error}</div>
  {:else if logs.length === 0}
    <div class="bg-[#12121e] border border-[#1a1a2e] rounded-xl p-16 text-center">
      <div class="text-slate-500 text-sm">No logs found</div>
    </div>
  {:else}
    <div class="bg-[#12121e] border border-[#1a1a2e] rounded-xl overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm" style="min-width: 600px">
          <thead>
            <tr class="border-b border-[#1a1a2e]">
              <th class="px-3 sm:px-4 py-3 text-left whitespace-nowrap"><button class="th-btn" on:click={() => setSort('created_at')}>Time{#if sortCol === 'created_at'} <span class="sort-icon">{sortDir === 'asc' ? '↑' : '↓'}</span>{/if}</button></th>
              <th class="px-3 sm:px-4 py-3 text-left"><button class="th-btn" on:click={() => setSort('model')}>Model{#if sortCol === 'model'} <span class="sort-icon">{sortDir === 'asc' ? '↑' : '↓'}</span>{/if}</button></th>
              <th class="px-3 sm:px-4 py-3 text-right"><button class="th-btn" on:click={() => setSort('input_tokens')}>In{#if sortCol === 'input_tokens'} <span class="sort-icon">{sortDir === 'asc' ? '↑' : '↓'}</span>{/if}</button></th>
              <th class="px-3 sm:px-4 py-3 text-right"><button class="th-btn" on:click={() => setSort('output_tokens')}>Out{#if sortCol === 'output_tokens'} <span class="sort-icon">{sortDir === 'asc' ? '↑' : '↓'}</span>{/if}</button></th>
              <th class="px-3 sm:px-4 py-3 text-right"><button class="th-btn" on:click={() => setSort('total_tokens')}>Total{#if sortCol === 'total_tokens'} <span class="sort-icon">{sortDir === 'asc' ? '↑' : '↓'}</span>{/if}</button></th>
              <th class="px-3 sm:px-4 py-3 text-right"><button class="th-btn" on:click={() => setSort('cost')}>Cost{#if sortCol === 'cost'} <span class="sort-icon">{sortDir === 'asc' ? '↑' : '↓'}</span>{/if}</button></th>
              <th class="px-3 sm:px-4 py-3 text-left"><span class="th-plain">Status</span></th>
            </tr>
          </thead>
          <tbody>
            {#each logs as log}
              <tr
                class="border-b border-[#1a1a2e]/40 hover:bg-[#1a1a2e]/30 transition-colors cursor-pointer"
                class:expanded={expandedId === log.id}
                on:click={() => toggleExpand(log.id)}
              >
                <td class="px-3 sm:px-4 py-2.5 text-slate-400 text-xs whitespace-nowrap">{fmtDate(log.created_at)}</td>
                <td class="px-3 sm:px-4 py-2.5">
                  <span class="font-mono text-xs text-blue-300 bg-blue-500/10 px-2 py-0.5 rounded">{log.model}</span>
                </td>
                <td class="px-3 sm:px-4 py-2.5 text-right text-slate-400 tabular-nums">{log.input_tokens.toLocaleString()}</td>
                <td class="px-3 sm:px-4 py-2.5 text-right text-slate-400 tabular-nums">{log.output_tokens.toLocaleString()}</td>
                <td class="px-3 sm:px-4 py-2.5 text-right text-slate-200 font-medium tabular-nums">{log.total_tokens.toLocaleString()}</td>
                <td class="px-3 sm:px-4 py-2.5 text-right text-slate-300 tabular-nums">{fmtCost(log.cost)}</td>
                <td class="px-3 sm:px-4 py-2.5">
                  {#if log.status === 'success'}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/15 text-emerald-400">ok</span>
                  {:else}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-500/15 text-red-400">err</span>
                  {/if}
                </td>
              </tr>
              {#if expandedId === log.id}
                <tr class="border-b border-[#1a1a2e]/40">
                  <td colspan="7" class="px-4 py-3 bg-[#0d0d14]">
                    <div class="space-y-2 text-xs">
                      {#if log.error}
                        <div>
                          <div class="text-red-400 font-medium mb-1">Error</div>
                          <div class="text-slate-400 font-mono bg-red-500/5 border border-red-500/20 rounded p-2">{log.error}</div>
                        </div>
                      {/if}
                      {#if parseMeta(log.metadata)}
                        <div>
                          <div class="text-slate-500 font-medium mb-1">Metadata</div>
                          <table class="bg-[#12121e] border border-[#1a1a2e] rounded overflow-hidden w-auto">
                            {#each Object.entries(parseMeta(log.metadata)) as [k, v]}
                              <tr class="border-b border-[#1a1a2e]/50 last:border-0">
                                <td class="px-3 py-1.5 text-slate-500 whitespace-nowrap pr-6">{k}</td>
                                <td class="px-3 py-1.5 text-slate-300">{metaVal(v)}</td>
                              </tr>
                            {/each}
                          </table>
                        </div>
                      {:else if !log.error}
                        <div class="text-slate-600">No additional details</div>
                      {/if}
                    </div>
                  </td>
                </tr>
              {/if}
            {/each}
          </tbody>
        </table>
      </div>

      {#if totalPages > 1}
        <div class="flex items-center justify-between px-4 py-3 border-t border-[#1a1a2e]">
          <span class="text-xs text-slate-500">Page {page} of {totalPages}</span>
          <div class="flex gap-2">
            <button on:click={prevPage} disabled={page === 1} class="btn-ghost text-xs disabled:opacity-40 disabled:cursor-not-allowed">Prev</button>
            <button on:click={nextPage} disabled={page >= totalPages} class="btn-ghost text-xs disabled:opacity-40 disabled:cursor-not-allowed">Next</button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .btn-primary {
    background: #4f46e5;
    color: white;
    font-size: 0.8125rem;
    padding: 0.375rem 0.75rem;
    border-radius: 0.5rem;
    transition: background 0.15s;
    border: none;
    cursor: pointer;
    white-space: nowrap;
  }
  .btn-primary:hover { background: #6366f1; }
  .btn-ghost {
    background: var(--raised);
    color: var(--text-muted);
    font-size: 0.8125rem;
    padding: 0.375rem 0.75rem;
    border-radius: 0.5rem;
    transition: all 0.15s;
    border: none;
    cursor: pointer;
    white-space: nowrap;
  }
  .btn-ghost:hover { background: var(--raised2); color: var(--text); }
  .th-btn {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
    transition: color 0.15s;
  }
  .th-btn:hover { color: var(--text-muted); }
  .th-plain {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .sort-icon { color: #6366f1; }
  .range-btn {
    background: none;
    border: 1px solid var(--border);
    border-radius: 0.375rem;
    padding: 0.25rem 0.625rem;
    font-size: 0.75rem;
    color: var(--text-dim);
    cursor: pointer;
    transition: all 0.15s;
  }
  .range-btn:hover { color: var(--text-muted); border-color: var(--raised2); }
  .range-btn.active {
    background: rgba(99, 102, 241, 0.15);
    border-color: rgba(99, 102, 241, 0.4);
    color: #a5b4fc;
  }
  tr.expanded > td { background: var(--bg); }
</style>
