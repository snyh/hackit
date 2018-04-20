import React, { Component } from 'react';

class PleaseUseClient extends Component {
    render() {
        return (
            <div>
                无法与本地服务器连接，请使用
                <a href="https://github.com/snyh/hackit/releases">服务端</a> 打开本页面。
                <a href="https://github.com/snyh/hackit">more</a>
            </div>
        );
    }
}

export {PleaseUseClient}
