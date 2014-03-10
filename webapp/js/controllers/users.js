angular.module('featen.user').controller('SignInController', ['$scope', '$routeParams', '$route', '$rootScope', '$location', 'Global', 'User', function($scope, $routeParams, $route, $rootScope, $location, Global, User) {
        $scope.data = {email: '', password: ''};
        $scope.signin = function() {
            User.signin({"Email": $scope.data.email, "Pass": $scope.data.password}, function(u) {
                Global.user = u;
                Global.authenticated = true;
                if (u.Type === 3 || u.Type === 0) {
                    Global.employee = u;
                    Global.isemployee=true;
                } 
//                else {
//                    Global.employee = null;
//                    Global.isemployee = false;
//                }
                $location.path('/');
            });
        };
    }]);

angular.module('featen.user').controller('SignInRedirectController', ['$scope', '$routeParams', '$route', '$rootScope', '$location', 'Global', 'User',"StageData", "Alerts", function($scope, $routeParams, $route, $rootScope, $location, Global, User, StageData, Alerts) {
    var redirectUrl = $routeParams.ReUrl;
    $scope.data = {email: '', password: ''};
    function redirect() {
        var enurl = redirectUrl;
        var deurl = Base64.decode(enurl);
        
        $location.path(deurl);
        $rootScope.$apply();
    };
    $scope.signin = function() {
        User.signin({"Email": $scope.data.email, "Pass": $scope.data.password}, function(u) {
            Global.user = u;
            Global.authenticated = true;
            if (u.Type === 3 || u.Type === 0) {
                Global.employee = u;
                Global.isemployee=true;
            }
            redirect();
        });
    };
}]);

angular.module('featen.user').controller('SignUpController', ['$scope', '$routeParams', '$location', 'Global', 'User', function($scope, $routeParams, $location, Global, User) {
        $scope.data = {email: '', password: ''};
        $scope.signup = function() {
            User.signup({"Email": $scope.data.email, "Pass": $scope.data.password}, function(u) {
                //Global.user = u;
                //Global.authenticated = true;
                $location.path('/signin');
            });
        };
    }]);

angular.module('featen.user').controller('RecoverController', ['$scope', '$routeParams', '$location', 'Global', 'User', function($scope, $routeParams, $location, Global, User) {
        $scope.data = {email: ''};
        $scope.sendRecoverMail = function() {
            User.sendRecoverMail({"Email": $scope.data.email}, function(u) {
                //Global.user = u;
                //Global.authenticated = true;
                $location.path('/');
            });
        };
    }]);


angular.module('featen.user').controller('UsersCtrl', ['$scope', '$routeParams', '$route', '$location', 'Global', 'User', 'Alerts', function($scope, $routeParams, $route, $location, Global, User, Alerts) {
        $scope.data = {nickname: '', password: '', passwordagain: '',phone:'',coverphoto:''};
        $scope.regaddr = {};
        $scope.shipaddr = {};
        var ordersTable;

        $scope.getCurrentUser = function() {
            User.get(function(data) {
                Global.user = data;
                Global.authenticated = true;
                if (data.Type === 3 || data.Type === 0) {
                    Global.employee = data;
                    Global.isemployee = true;
                } 
//                else {
//                    Global.employee = null;
//                    Global.isemployee = false;
//                }
                $scope.data.nickname = Global.user.Name;
                $scope.data.coverphoto = Global.user.CoverPhoto;
                $scope.data.phone = Global.user.Phone;
            });
        };

        //$scope.getCurrentUser();
        //$scope.global = Global;

        $scope.init = function() {
            $scope.getCurrentUser();            
            $scope.global = Global;
        };

        
        $scope.updatepassword = function() {
            User.updatepassword({"Pass": $scope.data.password}, function() {
                $location.path('/profile');
            	//$route.refresh();
            });
        };
        
        

        function uploadComplete(evt) {
            Alerts.add("success", "Success!", "Upload Photo Success");
            $scope.data.coverphoto = evt.currentTarget.responseText;
            $scope.$apply();
        };
        function uploadFailed(evt) {
            Alerts.add("error", "Failed!", "Upload failed, please check your file size and network。");
            $scope.$apply();
        }
        
        $scope.setFiles = function(element) {
            $scope.$apply(function($scope) {
                $scope.files = [];
                for (var i = 0; i < element.files.length; i++) {
                    $scope.files.push(element.files[i]);
                }
            });
        };

        $scope.uploadCoverPhoto = function() {
            var fd = new FormData();
            for (var i in $scope.files) {
                fd.append("files", $scope.files[i]);
            }
            var xhr = new XMLHttpRequest();
            xhr.addEventListener("load", uploadComplete, false);
            xhr.addEventListener("error", uploadFailed, false);
            xhr.open("POST", "/service/uploadphoto");
            xhr.send(fd);
        };

        $scope.updateinfo = function() {
            User.updateinfo({"Name": $scope.data.nickname, "Phone": $scope.data.phone, "CoverPhoto": $scope.data.coverphoto}, function() {
                //$location.path('/');
            	//$location.refresh();
            });
        };

        $scope.signout = function() {
            User.signout({"Id": $scope.global.user.id}, function() {
                Global.user = null;
                Global.employee=null;
                Global.authenticated = false;
                Global.isemployee = false;
                $location.path('/');
            });
        };

        $scope.getShippingOptions = function() {
            User.getShippingOptions(function(ss) {
            	if (ss.length == 2) {
	                if (ss[0].IsDefault === 1) {
	                    $scope.regaddr = ss[0];
	                    $scope.shipaddr = ss[1];
	                } else {
	                    $scope.regaddr = ss[1];
	                    $scope.shipaddr = ss[0];
	                }
            	} else if (ss.length == 1) {
            		if (ss[0].IsDefault === 1) {
            			$scope.regaddr = ss[0];
            		} else {
            			$scope.shipaddr = ss[0];
            		}
            	}
            	
            });
        };

        $scope.updateRegAddress = function() {
            $scope.regaddr.IsDefault = 1;
            User.updateAddress($scope.regaddr, function() {
                $location.path('/profile');
            });
        };

        $scope.updateShipAddress = function() {
            $scope.shipaddr.IsDefault = 0;
            User.updateAddress($scope.shipaddr, function() {
                $location.path('/profile');
            });
        };

        $scope.getOrders = function() {
            User.getUserOrders(function(ps) {
                $scope.orders = ps;
                if (ordersTable !== undefined && ordersTable !== null)
                    ordersTable.fnDestroy();
                $scope.initordertable();
            });
        };


        $scope.initordertable = function() {
            $("#orders-table tr").click(function() {
                $(this).toggleClass("row_selected");
            });
            ordersTable = $("#orders-table").dataTable({
                aaData: $scope.orders,
                aoColumns: [{
                        sTitle: "ID",
                        mData: "Id",
                        mRender: function(data, type, full) {
                            return '<a href="#!/order/' + data + '">' + data + "</a>";
                        }
                    }, {
                        sTitle: "Status",
                        mData: "Status", 
                        mRender: function(data, type, full) {
                            switch (data) {
                            case 0:
                                return "Placed";
                            case 1:
                                return "Paid";
                            case 2:
                                return "Delivered";
                            case 3:
                                return "Completed";
                            case 4:
                                return "Cancelled";
                            }
                        }
                    }, {
                        sTitle: "Paid",
                        mData: "PaidAmount"
                    }, {
                    	sTitle: "Created",
                    	mData: "CreateTime"
                    }]
            });
            $("table th input:checkBox").on("click", function() {
                var that = this;
                $(this).closest("table").find("tr > td:first-child input:checkbox").each(function() {
                    this.checked = that.checked;
                    $(this).closest("tr").toggleClass("selected");
                });
            });
        };



        $scope.getProfile = function() {
        	$scope.getCurrentUser();
            $scope.getShippingOptions();
            //$scope.getOrders();
        };


    
    }]);
