import {theme as antdTheme} from 'antd';

export const darkAntDesignTheme = {
    algorithm: antdTheme.darkAlgorithm,
    token: {
        colorPrimary: '#597ef7',
        colorInfo: '#86a3ff',
        colorBgBase: '#081a2f',
        colorBgContainer: 'rgba(0, 0, 0, 0.30)',
        colorBgElevated: '#17233d',
        colorBorder: 'rgba(255, 255, 255, 0.14)',
        colorSplit: 'rgba(255, 255, 255, 0.10)',
        colorText: 'rgba(255, 255, 255, 0.88)',
        colorTextSecondary: 'rgba(255, 255, 255, 0.64)',
        colorTextTertiary: 'rgba(255, 255, 255, 0.45)',
        borderRadius: 4,
        boxShadow: '0 12px 32px rgba(0, 0, 0, 0.28)',
    },
    components: {
        Layout: {
            bodyBg: 'transparent',
            headerBg: 'rgba(0, 0, 0, 0.30)',
        },
        Table: {
            headerBg: 'rgba(255, 255, 255, 0.08)',
            headerColor: 'rgba(255, 255, 255, 0.88)',
            rowHoverBg: 'rgba(89, 126, 247, 0.16)',
            borderColor: 'rgba(255, 255, 255, 0.10)',
        },
        Tabs: {
            itemColor: 'rgba(255, 255, 255, 0.66)',
            itemHoverColor: '#86a3ff',
            itemSelectedColor: '#ffffff',
            inkBarColor: '#597ef7',
        },
        Card: {
            colorBgContainer: 'rgba(0, 0, 0, 0.30)',
        },
        Modal: {
            contentBg: '#17233d',
            headerBg: '#17233d',
        },
    },
};
