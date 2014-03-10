angular.module('featen.enquire').controller('EnquiresController', ['$scope', '$routeParams', '$location', 'Global', 'Enquires', function ($scope, $routeParams, $location, Global, Enquires) {
    
    $scope.searchtext = "";
    $scope.searchcount = {};
    $scope.currPage = 1;
    $scope.totalPageNumber = 1;
    
    $scope.search = function() {
    	$scope.currPage = 1;
    	var t = $scope.searchtext;
    	if ($scope.searchtext.length == 0) 
    		t = " ";
    	Enquires.searchcount(t, function(sc) {
    		$scope.searchcount = sc;
    		$scope.totalPageNumber = Math.ceil(sc.Total / sc.PageLimit);
    	});
    	Enquires.search(t, 1, function(cs) {
    		$scope.enquires = cs;
    	});
    };
    
    $scope.setpage = function(n) {
    	var t = $scope.searchtext;
    	if ($scope.searchtext.length == 0) 
    		t = " ";
    	Enquires.search(t, n, function(cs) {
        	$scope.currPage = n;
    		$scope.enquires = cs;
    	});
    };
    
    $scope.getStatusName = function(s) {
        switch (s) {
        case 0:
            return "New";
        case 1:
            return "No Response";
        case 2:
            return "Completed";
        }
    };
}]);

angular.module('featen.enquire').controller('EnquireEditController', ['$scope', '$routeParams', '$location', 'Global', 'Enquires', function ($scope, $routeParams, $location, Global, Enquires) {
        
        var id = $routeParams.Id;
        $scope.get = function() {
        	$scope.data = {};	
            Enquires.getEnquire(id, function(o){
                $scope.data = o;
            });
        };
        
        $scope.addFollowup = function() {
        	var s = Number($scope.data.Status);
        	$scope.data.Status = s;
            Enquires.saveEnquire($scope.data, function(){
                $location.path("/enquires");
            });
        };
        
}]);
