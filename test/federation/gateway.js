const { ApolloServer } = require('apollo-server');
const { ApolloGateway } = require('@apollo/gateway');

const axios = require('axios');

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

var port = process.env.API_GQL_FEDERATION_PORT
var federatedServices = process.env.API_GQL_FEDERATION_SVCS

if (port === undefined) {
  console.error('The port to listen on MUST be specified via the API_GQL_FEDERATION_PORT env variable. Exiting.')
  process.exit()
}

if (federatedServices === undefined) {
  console.error('The services to federate MUST be specified via the API_GQL_FEDERATION_SVCS env variable. Exiting.')
  process.exit()
}

var serviceList = []

console.info("Parsing services to federate from '" + federatedServices + "'")

for (svc of federatedServices.split(",")) {
  console.info("Parsing svc config from: " + svc)

  svcConfig = svc.split(":")

  serviceList.push({
    name: svcConfig[0],
    url: "http://" + svcConfig[0] + ":" + svcConfig[1] + "/api/graphql"
  })
}

const gateway = new ApolloGateway({
    debug: true,
    serviceList: serviceList,
  });
  
  (async () => {
    try {
      // Start with checking if all services are available
      var serviceFailedToRespond = false
      for (service of serviceList) {
        await axios.get(service.url + "/playground")
                .then(response => {
                  console.info(service.name + " is reachable.");
                })
                .catch(error => {
                  serviceFailedToRespond = true
                });
        
        if (serviceFailedToRespond) {
          console.error(service.name + " is not reachable. Exiting!");
          await sleep(3000)
          process.exit()
        }
      }

      const { schema, executor } = await gateway.load();
  
      const server = new ApolloServer({ schema, executor });
    
      server.listen({port: port}).then(({ url }) => {
        console.log(`🚀 Server ready at ${url}`);
      });
    }
    catch(e)
    {
      console.error(e)
      // Wait 5 seconds before exiting to prevent hammering of the underlying services
      await sleep(5000);
      process.exit()
    }
  })();
