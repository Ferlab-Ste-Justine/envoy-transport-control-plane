'use strict';

const Hapi = require('hapi');

const runServer = (name, port) => {
    const server = new Hapi.Server();
    server.connection({ port: port });
    
    server.route([
        {
            method: 'GET',
            path: '/',
            handler: function (request, reply) {
                reply(name);
            }
        }
    ]);
    
    server.start((err) => {
        if (err) {
            throw err;
        }
        console.log('Server running at:', server.info.uri);
    });

    return server
}



const server1 = runServer("server1", 8081);
const server2 = runServer("server2", 8082);
const server3 = runServer("server3", 8083);
const server4 = runServer("server4", 8084);
const server5 = runServer("server5", 8085);


process.on('SIGTERM', () => {
    console.log('Sigterm received. Terminating server modules...');
    server1.stop((err) => {
        server2.stop((err) => {
            server3.stop((err) => {
                server4.stop((err) => {
                    server5.stop((err) => {
                        process.exit(0);
                    });
                });
            });
        });
    });
});