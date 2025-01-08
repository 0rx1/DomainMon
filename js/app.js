import { api } from './api.js';
import { initializeCharts, updateCharts } from './charts.js';

class DomainDashboard {
    constructor() {
        this.charts = initializeCharts();
        this.currentPage = 1;
        this.pageSize = 50;
        this.initializeEventListeners();
        this.loadDashboard();
        particlesJS('particles-js', {
            particles: {
                number: { value: 80, density: { enable: true, value_area: 800 } },
                color: { value: '#3b82f6' },
                opacity: { value: 0.1 },
                size: { value: 3 },
                line_linked: {
                    enable: true,
                    distance: 150,
                    color: '#3b82f6',
                    opacity: 0.1,
                    width: 1
                },
                move: {
                    enable: true,
                    speed: 2,
                    direction: 'none',
                    random: true,
                    out_mode: 'out'
                }
            }
        });
    }

    initializeEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-links li').forEach(link => {
            link.addEventListener('click', () => this.switchPage(link.dataset.page));
        });

        // Domain search
        const searchInput = document.getElementById('domainSearch');
        searchInput.addEventListener('input', this.debounce(() => this.loadDomains(), 300));

        // Lookup form
        document.getElementById('lookupBtn').addEventListener('click', () => this.performLookup());
    }

    async loadDashboard() {
        try {
            const [stats, domainsResponse] = await Promise.all([
                api.getDomainStats(),
                api.getDomains(1, 100)
            ]);

            if (stats && typeof stats === 'object') {
                this.updateStats(stats);
                if (domainsResponse && domainsResponse.domains) {
                    await updateCharts(this.charts, stats, domainsResponse.domains);
                }
            } else {
                console.error('Invalid stats data received');
                this.showError('Failed to load dashboard data');
            }
        } catch (error) {
            console.error('Failed to load dashboard:', error);
            this.showError('Failed to connect to the API');
        }
    }

    async loadDomains() {
        try {
            const search = document.getElementById('domainSearch').value;
            const data = await api.getDomains(this.currentPage, this.pageSize, search);
            this.renderDomainsTable(data);
        } catch (error) {
            console.error('Failed to load domains:', error);
        }
    }

    async performLookup() {
        const domain = document.getElementById('lookupDomain').value;
        if (!domain) return;

        try {
            const [whois, dns] = await Promise.all([
                api.getWhoisInfo(domain),
                api.getDNSInfo(domain)
            ]);

            this.renderWhoisInfo(whois);
            this.renderDNSInfo(dns);
        } catch (error) {
            console.error('Lookup failed:', error);
        }
    }

    updateStats(stats) {
        if (!stats) return;
        
        const totalDomains = document.getElementById('totalDomains');
        const totalTLDs = document.getElementById('totalTLDs');
        const onlineDomains = document.getElementById('onlineDomains');

        if (totalDomains) totalDomains.textContent = stats.total_domains || 0;
        if (totalTLDs) totalTLDs.textContent = stats.domains_per_tld ? Object.keys(stats.domains_per_tld).length : 0;
        if (onlineDomains) onlineDomains.textContent = '-'; // Will be updated when we have health data
    }

    renderDomainsTable(data) {
        const tbody = document.getElementById('domainsTableBody');
        tbody.innerHTML = data.domains.map(domain => `
            <tr>
                <td>${domain.name}</td>
                <td>${domain.tld}</td>
                <td class="actions">
                    <button class="health-btn" data-domain="${domain.name}" title="Check health">
                        <i class="fas fa-heartbeat"></i>
                    </button>
                    <button class="lookup-btn" data-domain="${domain.name}" title="Lookup domain">
                        <i class="fas fa-search"></i>
                    </button>
                </td>
                <td>${new Date(domain.created_at).toLocaleDateString()}</td>
            </tr>
        `).join('');

        // Add event listeners to buttons
        tbody.querySelectorAll('.lookup-btn').forEach(button => {
            button.addEventListener('click', () => {
                const domain = button.dataset.domain;
                this.switchPage('lookups');
                document.getElementById('lookupDomain').value = domain;
                this.performLookup();
            });
        });

        tbody.querySelectorAll('.health-btn').forEach(button => {
            button.addEventListener('click', async () => {
                const domain = button.dataset.domain;
                button.classList.add('loading');
                try {
                    const health = await api.getDomainHealth(domain);
                    this.showHealthStatus(button, health);
                } catch (error) {
                    console.error('Health check failed:', error);
                    this.showHealthStatus(button, { is_online: false, error: 'Check failed' });
                }
            });
        });
    }

    showHealthStatus(button, health) {
        button.classList.remove('loading');
        const row = button.closest('tr');
        
        // Remove any existing status
        const existingStatus = row.querySelector('.status-popup');
        if (existingStatus) {
            existingStatus.remove();
        }

        // Create and show new status
        const status = document.createElement('div');
        status.className = `status-popup ${health.is_online ? 'online' : 'offline'}`;
        status.innerHTML = `
            <div class="status-header">
                ${health.is_online ? 'Online' : 'Offline'}
            </div>
            ${health.protocol ? `<div>Protocol: ${health.protocol}</div>` : ''}
            ${health.status_code ? `<div>Status: ${health.status_code}</div>` : ''}
            ${health.response_time ? `<div>Response: ${health.response_time}</div>` : ''}
            ${health.error ? `<div class="error">${health.error}</div>` : ''}
        `;

        // Position the popup next to the button
        button.parentNode.appendChild(status);

        // Remove popup after 5 seconds
        setTimeout(() => status.remove(), 5000);
    }

    renderWhoisInfo(whois) {
        if (!whois) return;

        const content = document.querySelector('#whoisResult .result-content');
        content.innerHTML = `
            <div class="info-grid">
                <div class="info-item">
                    <label>Registrar</label>
                    <span>${whois.registrar || 'N/A'}</span>
                </div>
                <div class="info-item">
                    <label>Created Date</label>
                    <span>${whois.created_date ? new Date(whois.created_date).toLocaleDateString() : 'N/A'}</span>
                </div>
                <div class="info-item">
                    <label>Expiry Date</label>
                    <span>${whois.expiry_date ? new Date(whois.expiry_date).toLocaleDateString() : 'N/A'}</span>
                </div>
                <div class="info-item">
                    <label>DNSSEC</label>
                    <span>${whois.dnssec ? 'Yes' : 'No'}</span>
                </div>
                <div class="info-item">
                    <label>Nameservers</label>
                    <span>${(whois.nameservers || []).join(', ') || 'N/A'}</span>
                </div>
            </div>
        `;

        if (whois.error) {
            content.innerHTML += `
                <div class="error-message mt-4">
                    ${whois.error}
                </div>
            `;
        }
    }

    renderDNSInfo(dns) {
        if (!dns) return;

        const content = document.querySelector('#dnsResult .result-content');
        content.innerHTML = `
            <div class="info-grid">
                <div class="info-item">
                    <label>IP Addresses</label>
                    ${(dns.ip_addresses || []).map(ip => `
                        <div>
                            ${ip.address} (IPv${ip.version})
                            ${ip.reverse ? `<div class="text-sm">Reverse: ${ip.reverse.join(', ')}</div>` : ''}
                        </div>
                    `).join('')}
                </div>
                ${(dns.mx_records || []).length ? `
                    <div class="info-item">
                        <label>MX Records</label>
                        ${dns.mx_records.map(mx => `
                            <div>${mx.host} (Priority: ${mx.priority})</div>
                        `).join('')}
                    </div>
                ` : ''}
                ${(dns.txt_records || []).length ? `
                    <div class="info-item">
                        <label>TXT Records</label>
                        ${dns.txt_records.map(txt => `<div class="text-sm">${txt}</div>`).join('')}
                    </div>
                ` : ''}
                ${(dns.ns_records || []).length ? `
                    <div class="info-item">
                        <label>NS Records</label>
                        ${dns.ns_records.map(ns => `
                            <div>
                                ${ns.host}
                                ${ns.ips ? `<div class="text-sm">IPs: ${ns.ips.join(', ')}</div>` : ''}
                            </div>
                        `).join('')}
                    </div>
                ` : ''}
            </div>
        `;

        if (dns.error) {
            content.innerHTML += `
                <div class="error-message mt-4">
                    ${dns.error}
                </div>
            `;
        }
    }

    switchPage(pageId) {
        document.querySelectorAll('.page').forEach(page => {
            page.classList.remove('active');
        });
        document.getElementById(pageId).classList.add('active');

        document.querySelectorAll('.nav-links li').forEach(link => {
            link.classList.toggle('active', link.dataset.page === pageId);
        });

        if (pageId === 'domains') {
            this.loadDomains();
        }
    }

    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    showError(message) {
        // Add error display functionality
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;
        document.querySelector('.content').prepend(errorDiv);
    }
}

class ThemeSwitcher {
    static themes = {
        light: {
            '--background-color': '#f3f4f6',
            '--text-color': '#1f2937',
            '--card-background': '#ffffff'
        },
        dark: {
            '--background-color': '#1f2937',
            '--text-color': '#f3f4f6',
            '--card-background': '#374151'
        }
    };

    static switchTheme(theme) {
        const root = document.documentElement;
        Object.entries(this.themes[theme]).forEach(([property, value]) => {
            root.style.setProperty(property, value);
        });
    }
}

// Initialize the dashboard
new DomainDashboard(); 