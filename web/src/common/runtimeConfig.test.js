import test from 'node:test';
import assert from 'node:assert/strict';

import {getInitialLocale, getThemeClassName, isRuntimeFlagEnabled} from './runtimeConfig.js';

test('isRuntimeFlagEnabled accepts common true values', () => {
    for (const value of [true, 'true', ' TRUE ', '1', 'yes', 'on']) {
        assert.equal(isRuntimeFlagEnabled(value), true);
    }
});

test('isRuntimeFlagEnabled rejects disabled values', () => {
    for (const value of [false, '', 'false', '0', 'no', 'off', undefined, null]) {
        assert.equal(isRuntimeFlagEnabled(value), false);
    }
});

test('getInitialLocale forces zh-cn in iframe mode', () => {
    const locale = getInitialLocale({
        iframeMode: true,
        storedLocale: 'en-us',
    });

    assert.equal(locale, 'zh-cn');
});

test('getInitialLocale uses stored locale outside iframe mode', () => {
    const locale = getInitialLocale({
        iframeMode: false,
        storedLocale: 'en-us',
    });

    assert.equal(locale, 'en-us');
});

test('getThemeClassName includes the dark theme hook when enabled', () => {
    assert.equal(getThemeClassName(true), 'km-app km-app-dark');
});

test('getThemeClassName leaves the default theme unmodified', () => {
    assert.equal(getThemeClassName(false), 'km-app');
});
