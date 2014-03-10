function AlertsCtrl($scope, Alerts) {
		$scope.alerts = Alerts.alerts;
		
		$scope.remove = function(index) {
				Alerts.remove(index);
		};
}
AlertsCtrl.$inject = ['$scope', 'Alerts'];
