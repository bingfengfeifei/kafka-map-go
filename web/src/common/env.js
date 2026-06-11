function env() {
    const joinPath = (...parts) => {
        const filtered = parts.filter(part => part !== undefined && part !== null && part !== '');
        if (filtered.length === 0) {
            return '';
        }
        return '/' + filtered.map(part => String(part).replace(/^\/+|\/+$/g, '')).filter(Boolean).join('/');
    };

    const currentBasePath = () => {
        const viteBase = import.meta.env.BASE_URL || '/';
        if (viteBase !== './' && viteBase !== '/') {
            return viteBase;
        }
        return window.location.pathname;
    };

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
        const apiBase = import.meta.env.VITE_API_BASE || joinPath(basePath, 'api');
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

export const appBasePath = function () {
    const viteBase = import.meta.env.BASE_URL || '/';
    if (viteBase !== './' && viteBase !== '/') {
        return viteBase.endsWith('/') ? viteBase : viteBase + '/';
    }
    return '/';
}();
