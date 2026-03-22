import { authStore } from './stores.js';
import { get } from 'svelte/store';

class API {
  async request(method, url, body) {
    const { token } = get(authStore);
    const headers = { 'Content-Type': 'application/json' };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const res = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || 'Request failed');
    }

    return data;
  }

  get(url) {
    return this.request('GET', url);
  }

  post(url, body) {
    return this.request('POST', url, body);
  }

  del(url) {
    return this.request('DELETE', url);
  }
}

export const api = new API();
