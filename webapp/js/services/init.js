var services = angular.module('services', []);

services.service('Alerts', ['$rootScope', AlertsService]);
services.service('User', ['$http', 'Alerts', UsersService]);
services.service('Articles', ['$http', 'Alerts', ArticlesService]);
services.service('Msb', ['$http', 'Alerts', MsbService]);
services.service('PerfSimu', ['$http', 'Alerts', PerfSimuService]);
