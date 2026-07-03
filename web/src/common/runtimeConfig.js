export const getRuntimeConfig = () => {
    if (typeof window === 'undefined' || !window.__KAFKA_MAP_CONFIG__) {
        return {};
    }
    return window.__KAFKA_MAP_CONFIG__;
};

export const isRuntimeFlagEnabled = (value) => {
    if (typeof value === 'boolean') {
        return value;
    }
    if (typeof value === 'number') {
        return value === 1;
    }
    if (typeof value !== 'string') {
        return false;
    }

    switch (value.trim().toLowerCase()) {
        case '1':
        case 'true':
        case 'yes':
        case 'on':
            return true;
        default:
            return false;
    }
};

export const getInitialLocale = ({iframeMode = false, storedLocale, fallbackLocale = 'en-us'} = {}) => {
    if (iframeMode) {
        return 'zh-cn';
    }
    return storedLocale || fallbackLocale;
};

export const getThemeClassName = (darkTheme = false) => {
    return darkTheme ? 'km-app km-app-dark' : 'km-app';
};
