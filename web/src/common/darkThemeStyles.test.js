import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const projectRoot = path.resolve(import.meta.dirname, '../../');

const readSource = (relativePath) => {
    return fs.readFileSync(path.join(projectRoot, relativePath), 'utf8');
};

const extractPopconfirmBlocks = (source) => {
    const blocks = [];
    const pattern = /<Popconfirm\b[\s\S]*?<\/Popconfirm>/g;
    let match;
    while ((match = pattern.exec(source)) !== null) {
        blocks.push(match[0]);
    }
    return blocks;
};

test('Popconfirm actions use localized confirm and cancel text', () => {
    const files = [
        'src/components/Cluster.jsx',
        'src/components/ConsumerGroup.jsx',
        'src/components/Topic.jsx',
    ];

    for (const file of files) {
        for (const block of extractPopconfirmBlocks(readSource(file))) {
            assert.match(block, /okText=\{<FormattedMessage id="okText"\s*\/>\}/, `${file} Popconfirm missing localized okText`);
            assert.match(block, /cancelText=\{<FormattedMessage id="cancelText"\s*\/>\}/, `${file} Popconfirm missing localized cancelText`);
        }
    }
});

test('dark modal portal styles cover close button, labels, and form controls', () => {
    const css = readSource('src/App.css');

    assert.match(css, /\.km-body-dark\s+\.ant-modal-close/, 'missing dark modal close button style');
    assert.match(css, /\.km-body-dark\s+\.ant-form-item-label\s*>\s*label/, 'missing dark modal label style');
    assert.match(css, /\.km-body-dark\s+\.ant-input/, 'missing dark modal input style');
    assert.match(css, /\.km-body-dark\s+\.ant-select-selector/, 'missing dark modal select style');
});
