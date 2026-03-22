<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api.js';

  let daily = [];
  let loading = true;
  let error = null;
  let days = 30;

  const dayOptions = [7, 14, 30, 60, 90];

  async function load() {
    loading = true;
    error = null;
    try {
      daily = await api.get(`/api/stats/daily?days=${days}`);
    } catch (e) {
      error = e.message;
    }
    loading = false;
  }

  onMount(load);

  function selectDays(d) {
    days = d;
    load();
  }

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

  $: maxCount = Math.max(...daily.map(d => d.count), 1);
  $: maxCost = Math.max(...daily.map(d => d.cost), 0.0001);
  $: maxTokens = Math.max(...daily.map(d => d.tokens), 1);
  $: maxLatency = Math.max(...daily.map(d => d.avg_latency_ms), 1);

  $: totalCost = daily.reduce((s, d) => s + d.cost, 0);
  $: totalRequests = daily.reduce((s, d) => s + d.count, 0);
  $: totalTokens = daily.reduce((s, d) => s + d.tokens, 0);
  $: avgLatency = daily.length ? daily.reduce((s, d) => s + d.avg_latency_ms, 0) / daily.length : 0;
</script>

<div class="p-4 sm:p-6 max-w-6xl mx-auto">
  <div class="flex items-center justify-between mb-5">
    <div>
      <h1 class="text-xl font-semibold text-white">Stats</h1>
      <p class="text-slate-500 text-sm mt-0.5">Detailed usage charts</p>
    </div>
    <div class="flex gap-1">
      {#each dayOptions as d}
        <button
          class="day-btn"
          class:active={days === d}
          on:click={() => selectDays(d)}
        >
          {d}d
        </button>
      {/each}
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-24">
      <div class="w-7 h-7 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if error}
    <div class="bg-red-500/10 border border-red-500/30 rounded-xl p-4 text-red-400 text-sm">{error}</div>
  {:else if daily.length === 0}
    <div class="panel text-center py-16">
      <div class="text-slate-500 text-sm">No data for the last {days} days.</div>
    </div>
  {:else}
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-5">
      <div class="card">
        <div class="card-label">Requests</div>
        <div class="card-value">{fmt(totalRequests)}</div>
        <div class="card-sub">last {days} days</div>
      </div>
      <div class="card">
        <div class="card-label">Tokens</div>
        <div class="card-value">{fmt(totalTokens)}</div>
        <div class="card-sub">last {days} days</div>
      </div>
      <div class="card">
        <div class="card-label">Cost</div>
        <div class="card-value">{fmtCost(totalCost)}</div>
        <div class="card-sub">last {days} days</div>
      </div>
      <div class="card">
        <div class="card-label">Avg Latency</div>
        <div class="card-value">{fmtMs(avgLatency)}</div>
        <div class="card-sub">last {days} days</div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- Cost Chart -->
      <div class="panel">
        <div class="text-xs font-medium text-slate-400 mb-3 uppercase tracking-wide">Daily Cost</div>
        <div class="chart-wrap">
          {#each daily as day}
            <div class="chart-col group">
              <div class="chart-tooltip">
                {day.date}<br/>{fmtCost(day.cost)}
              </div>
              <div
                class="chart-bar bar-green"
                style="height: {Math.max((day.cost / maxCost) * 100, 3)}%"
              ></div>
            </div>
          {/each}
        </div>
        <div class="chart-labels">
          <span>{daily[0]?.date}</span>
          <span>{daily[daily.length - 1]?.date}</span>
        </div>
      </div>

      <!-- Requests Chart -->
      <div class="panel">
        <div class="text-xs font-medium text-slate-400 mb-3 uppercase tracking-wide">Daily Requests</div>
        <div class="chart-wrap">
          {#each daily as day}
            <div class="chart-col group">
              <div class="chart-tooltip">
                {day.date}<br/>{day.count} req
              </div>
              <div
                class="chart-bar bar-indigo"
                style="height: {Math.max((day.count / maxCount) * 100, 3)}%"
              ></div>
            </div>
          {/each}
        </div>
        <div class="chart-labels">
          <span>{daily[0]?.date}</span>
          <span>{daily[daily.length - 1]?.date}</span>
        </div>
      </div>

      <!-- Tokens Chart -->
      <div class="panel">
        <div class="text-xs font-medium text-slate-400 mb-3 uppercase tracking-wide">Daily Tokens</div>
        <div class="chart-wrap">
          {#each daily as day}
            <div class="chart-col group">
              <div class="chart-tooltip">
                {day.date}<br/>{fmt(day.tokens)} tokens
              </div>
              <div
                class="chart-bar bar-amber"
                style="height: {Math.max((day.tokens / maxTokens) * 100, 3)}%"
              ></div>
            </div>
          {/each}
        </div>
        <div class="chart-labels">
          <span>{daily[0]?.date}</span>
          <span>{daily[daily.length - 1]?.date}</span>
        </div>
      </div>

      <!-- Latency Chart -->
      <div class="panel">
        <div class="text-xs font-medium text-slate-400 mb-3 uppercase tracking-wide">Avg Latency</div>
        <div class="chart-wrap">
          {#each daily as day}
            <div class="chart-col group">
              <div class="chart-tooltip">
                {day.date}<br/>{fmtMs(day.avg_latency_ms)}
              </div>
              <div
                class="chart-bar bar-rose"
                style="height: {Math.max((day.avg_latency_ms / maxLatency) * 100, 3)}%"
              ></div>
            </div>
          {/each}
        </div>
        <div class="chart-labels">
          <span>{daily[0]?.date}</span>
          <span>{daily[daily.length - 1]?.date}</span>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .day-btn {
    padding: 0.25rem 0.625rem;
    border-radius: 0.375rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-muted);
    background: var(--surface);
    border: 1px solid var(--border);
    cursor: pointer;
    transition: all 0.15s;
  }
  .day-btn:hover {
    color: var(--text);
    background: var(--raised);
  }
  .day-btn.active {
    color: white;
    background: rgba(99, 102, 241, 0.8);
    border-color: rgba(99, 102, 241, 0.6);
  }

  .chart-wrap {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 120px;
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
    border-radius: 2px 2px 0 0;
    transition: opacity 0.15s;
    cursor: default;
  }
  .chart-col:hover .chart-bar {
    opacity: 1;
  }

  .bar-indigo { background: rgba(99, 102, 241, 0.7); }
  .chart-col:hover .bar-indigo { background: #6366f1; }

  .bar-green { background: rgba(34, 197, 94, 0.7); }
  .chart-col:hover .bar-green { background: #22c55e; }

  .bar-amber { background: rgba(245, 158, 11, 0.7); }
  .chart-col:hover .bar-amber { background: #f59e0b; }

  .bar-rose { background: rgba(244, 63, 94, 0.7); }
  .chart-col:hover .bar-rose { background: #f43f5e; }

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

  .chart-labels {
    display: flex;
    justify-content: space-between;
    font-size: 0.75rem;
    color: var(--text-muted);
    margin-top: 0.375rem;
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
