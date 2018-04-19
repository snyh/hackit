import ReactDOM from 'react-dom';
import 'semantic-ui-css/semantic.min.css';
import './index.css';
import './App.css';

import { BrowserRouter } from 'react-router-dom';
import registerServiceWorker from './registerServiceWorker';
import React, { Component } from 'react';
import { Container, Segment } from 'semantic-ui-react';
import { Route, Switch } from 'react-router-dom';
import UserView from './UserView.js';
import HackerView from './HackView.js';
import { PleaseUseClient } from './Widget.js';

class App extends Component {
    render() {
        return (
            <Container>
                <Switch>
                    <Route exact path="/" component={PleaseUseClient}/>
                    <Route path="/connect/:id" component={HackerView}/>
                    <Route path="/mysys/:port" component={UserView} />
                </Switch>
                <Segment.Group horizontal class=".App-footer">
                    <Segment>
                        <a href="https://github.com/snyh/hackit/releases">客户端下载</a>
                    </Segment>

                    <Segment>
                        <a href="https://github.com/snyh/hackit">源码</a>
                    </Segment>
                </Segment.Group>
            </Container>
        );
    }
}

ReactDOM.render(
    <BrowserRouter>
        <App />
    </BrowserRouter>,
    document.getElementById('root')
);


registerServiceWorker();
