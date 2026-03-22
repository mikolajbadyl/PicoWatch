import { writable } from 'svelte/store';

export const authStore = writable({
  token: null,
  authenticated: false,
});
