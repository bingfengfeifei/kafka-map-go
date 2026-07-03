const joinPath = (...parts) => {
    const filtered = parts.filter(part => part !== undefined && part !== null && part !== '');
    if (filtered.length === 0) {
        return '';
    }
    return '/' + filtered.map(part => String(part).replace(/^\/+|\/+$/g, '')).filter(Boolean).join('/');
};

const normalizeBasePath = (basePath) => {
    const value = String(basePath || '').trim();
    if (!value || value === '/' || value === './') {
        return '/';
    }

    const prefixed = value.startsWith('/') ? value : `/${value}`;
    return prefixed.endsWith('/') ? prefixed : `${prefixed}/`;
};

const runtimeConfig = () => window.__KAFKA_MAP_CONFIG__ || {};

const currentBasePath = () => {
    const configBase = runtimeConfig().basePath;
    if (configBase) {
        return normalizeBasePath(configBase);
    }

    const viteBase = import.meta.env.BASE_URL || '/';
    if (viteBase !== './' && viteBase !== '/') {
        return normalizeBasePath(viteBase);
    }

    return normalizeBasePath(window.location.pathname);
};

function env() {
    if (import.meta.env.MODE === 'development') {
        // 本地开发环境
        return {
            server: '//127.0.0.1:8080',
            wsServer: 'ws://127.0.0.1:8080',
            prefix: '',
        }
    } else {
        // 生产环境
        let wsPrefix;
        if (window.location.protocol === 'https:') {
            wsPrefix = 'wss:'
        } else {
            wsPrefix = 'ws:'
        }
        const basePath = currentBasePath();
        const apiBase = runtimeConfig().apiBase || import.meta.env.VITE_API_BASE || joinPath(basePath, 'api');
        return {
            server: apiBase,
            wsServer: wsPrefix + window.location.host,
            prefix: window.location.protocol + '//' + window.location.host,
        }
    }
}
export default env();

export const server = env().server;
export const wsServer = env().wsServer;
export const prefix = env().prefix;

export const appBasePath = currentBasePath();

export const authDisabled = !!runtimeConfig().authDisabled;
