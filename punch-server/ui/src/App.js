import React, { Component } from 'react';
import './App.css';

import { Container, Header, Divider } from 'semantic-ui-react';

import { Route, Link, Switch } from 'react-router-dom';

class MagicLink extends Component {
    render() {
        const id = this.props.match.params.id
        return <div>Hello... try connecting to {id}</div>
    }
}

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
            return <li key={id}><Link to={`/connect/${id}`}>{id}</Link></li>
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
                <Header> HackIt 管理后台 <Link to="/">Home</Link> </Header>
                <Divider/>
                <Switch>
                    <Route exact path="/" component={ListMagicLink}/>
                    <Route path="/connect/:id" component={MagicLink}/>
                </Switch>
            </Container>
        );
    }
}

export default App;
