const API_BASE_URL = 'http://localhost:8080/api/v1';

export async function fetchJson(endpoint, options = {}) {
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            ...options,
            headers: {
                'Accept': 'application/json',
                ...options.headers,
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        return data;
    } catch (error) {
        console.error('API Error:', error);
        return null;
    }
}

export const api = {
    async getDomainStats() {
        const data = await fetchJson('/domains/stats');
        if (!data) return { total_domains: 0, domains_per_tld: {} };
        return data;
    },

    async getDomains(page = 1, limit = 50, search = '') {
        const data = await fetchJson(`/domains?page=${page}&limit=${limit}&search=${search}`);
        if (!data) return { domains: [], total: 0, page: 1, limit };
        return data;
    },

    async getWhoisInfo(domain) {
        return await fetchJson(`/lookup/whois?domain=${domain}`);
    },

    async getDNSInfo(domain) {
        return await fetchJson(`/lookup/dns?domain=${domain}`);
    },

    async getDomainHealth(domain) {
        return await fetchJson(`/domains/health?domain=${domain}`);
    }
}; 