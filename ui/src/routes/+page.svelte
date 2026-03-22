<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';

  let stats = null;
  let models = [];
  let daily = [];
  let loading = true;
  let error = null;

  onMount(async () => {
    try {
      [stats, models, daily] = await Promise.all([
        api.get('/api/stats'),
        api.get('/api/stats/models'),
        api.get('/api/stats/daily'),
      ]);
    } catch (e) {
      error = e.message;
    }
    loading = false;
  });

  function fmt(n) {
    if (n == null) return '0';
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
    return Math.round(n).toString();
  }

  function fmtCost(n) {
    return '$' + (n ?? 0).toFixed(4);
  }

  function fmtMs(n) {
    if (!n) return '0ms';
    if (n >= 1000) return (n / 1000).toFixed(1) + 's';
    return Math.round(n) + 'ms';
  }

  $: maxDaily = Math.max(...daily.map((d) => d.count), 1);
</script>

<div class="p-4 sm:p-6 max-w-6xl mx-auto">
  <div class="mb-5">
    <h1 class="text-xl font-semibold text-white">Overview</h1>
    <p class="text-slate-500 text-sm mt-0.5">LLM usage analytics</p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-24">
      <div class="w-7 h-7 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if error}
    <div class="bg-red-500/10 border border-red-500/30 rounded-xl p-4 text-red-400 text-sm">{error}</div>
  {:else}
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-5">
      <div class="card">
        <div class="card-label">Requests</div>
        <div class="card-value">{fmt(stats.total_requests)}</div>
        <div class="card-sub">{stats.requests_today} today</div>
      </div>
      <div class="card">
        <div class="card-label">Tokens</div>
        <div class="card-value">{fmt(stats.total_tokens)}</div>
        <div class="card-sub">{fmt(stats.total_input_tokens)} in / {fmt(stats.total_output_tokens)} out</div>
      </div>
      <div class="card">
        <div class="card-label">Total Cost</div>
        <div class="card-value">{fmtCost(stats.total_cost)}</div>
        <div class="card-sub">{fmtCost(stats.cost_month)} this month</div>
      </div>
      <div class="card">
        <div class="card-label">Today</div>
        <div class="card-value">{fmtCost(stats.cost_today)}</div>
        <div class="card-sub">{stats.requests_today} requests</div>
      </div>
    </div>

    {#if daily.length > 0}
      <div class="panel mb-4">
        <div class="text-xs font-medium text-slate-400 mb-3 uppercase tracking-wide">Requests — last 30 days</div>
        <div class="chart-wrap">
          {#each daily as day}
            <div class="chart-col group">
              <div class="chart-tooltip">
                {day.date}<br/>{day.count} req
              </div>
              <div
                class="chart-bar"
                style="height: {Math.max((day.count / maxDaily) * 100, 3)}%"
              ></div>
            </div>
          {/each}
        </div>
        <div class="flex justify-between text-xs text-slate-600 mt-1.5">
          <span>{daily[0]?.date}</span>
          <span>{daily[daily.length - 1]?.date}</span>
        </div>
      </div>
    {/if}

    {#if models.length > 0}
      <div class="panel">
        <div class="text-xs font-medium text-slate-400 mb-4 uppercase tracking-wide">Models</div>
        <div class="overflow-x-auto -mx-1">
          <table class="w-full text-sm min-w-[480px]">
            <thead>
              <tr class="text-left border-b border-[#1a1a2e]">
                <th class="pb-2.5 text-xs font-medium text-slate-500 px-1">Model</th>
                <th class="pb-2.5 text-xs font-medium text-slate-500 text-right px-1">Requests</th>
                <th class="pb-2.5 text-xs font-medium text-slate-500 text-right px-1 hidden sm:table-cell">Input</th>
                <th class="pb-2.5 text-xs font-medium text-slate-500 text-right px-1 hidden sm:table-cell">Output</th>
                <th class="pb-2.5 text-xs font-medium text-slate-500 text-right px-1">Cost</th>
                <th class="pb-2.5 text-xs font-medium text-slate-500 text-right px-1">Latency</th>
              </tr>
            </thead>
            <tbody>
              {#each models as model}
                <tr class="border-b border-[#1a1a2e]/40 hover:bg-[#1a1a2e]/30 transition-colors">
                  <td class="py-2.5 pr-3 px-1">
                    <span class="font-mono text-xs text-blue-300 bg-blue-500/10 px-2 py-0.5 rounded">{model.model}</span>
                  </td>
                  <td class="py-2.5 text-right text-slate-300 px-1">{model.count.toLocaleString()}</td>
                  <td class="py-2.5 text-right text-slate-400 px-1 hidden sm:table-cell">{fmt(model.input_tokens)}</td>
                  <td class="py-2.5 text-right text-slate-400 px-1 hidden sm:table-cell">{fmt(model.output_tokens)}</td>
                  <td class="py-2.5 text-right text-slate-300 px-1">{fmtCost(model.total_cost)}</td>
                  <td class="py-2.5 text-right text-slate-400 px-1">{fmtMs(model.avg_latency_ms)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {:else}
      <div class="panel text-center py-16">
        <div class="text-slate-500 text-sm">No data yet.</div>
        <div class="text-slate-600 text-xs mt-2">Start logging via <span class="font-mono text-blue-400">POST /api/logs</span></div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .chart-wrap {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 80px;
  }
  .chart-col {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: flex-end;
    position: relative;
  }
  .chart-bar {
    width: 100%;
    background: rgba(99, 102, 241, 0.7);
    border-radius: 2px 2px 0 0;
    transition: background 0.15s;
    cursor: default;
  }
  .chart-col:hover .chart-bar {
    background: #6366f1;
  }
  .chart-tooltip {
    display: none;
    position: absolute;
    bottom: calc(100% + 6px);
    left: 50%;
    transform: translateX(-50%);
    background: var(--raised);
    border: 1px solid var(--raised2);
    border-radius: 4px;
    padding: 3px 7px;
    font-size: 0.7rem;
    color: var(--text);
    white-space: nowrap;
    z-index: 10;
    pointer-events: none;
    text-align: center;
    line-height: 1.4;
  }
  .chart-col:hover .chart-tooltip {
    display: block;
  }
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    padding: 1rem;
  }
  .card-label {
    font-size: 0.7rem;
    font-weight: 500;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.375rem;
  }
  .card-value {
    font-size: 1.625rem;
    font-weight: 700;
    color: var(--text-heading);
    line-height: 1;
  }
  .card-sub {
    font-size: 0.7rem;
    color: var(--text-muted);
    margin-top: 0.375rem;
  }
  .panel {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    padding: 1.25rem;
  }
</style>
