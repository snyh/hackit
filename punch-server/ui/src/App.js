import React, { Component } from 'react';
import './App.css';

import { Container, Header, Divider } from 'semantic-ui-react';

class ListMagicLink extends Component {
    constructor(props) {
        super(props)
        this.state = {
            values : []
        }
    }

    componentDidMount() {
        const tick = () => {
            fetch("http://localhost:2207/list").then( (resp) => {
                resp.json().then( data => {
                    this.setState({values: data});
                    this.timer = setTimeout(tick, 3000);
                })
            })
        };
        tick();
    }

    componentWillUnmount() {
        clearTimeout(this.timer);
    }

    render() {
        const ids = this.state.values.map( id => {
            return <li key={id}><a target="_blank" href={`/connect/${id}`}>{id}</a></li>
        })
        return (
            <div>
                <Header> 当前有 {ids.length} 个连接 </Header>
                <ul>{ids}</ul>
            </div>
        );
    }
}

class App extends Component {
    render() {
        return (
            <Container>
                <Header> HackIt 管理后台 </Header>
                <Divider/>
                <ListMagicLink values={["12345", "54321"]} />
            </Container>
        );
    }
}

export default App;
