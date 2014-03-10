//Setting up route
window.app.config(['$routeProvider',
    function($routeProvider) {
        $routeProvider.
                when('/signin', {
                    templateUrl: 'views/signin.html'
                }).
                when('/signin/redirect/:ReUrl', {
                    templateUrl: 'views/signin_redirect.html'
                }).
                when('/signup', {
                    templateUrl: 'views/signup.html'
                }).
                when('/recover', {
                    templateUrl: 'views/recover.html'
                }).
                when('/profile', {
                    templateUrl: 'views/profile.html'
                }).
                when('/myinfo', {
                    templateUrl: 'views/myinfo.html'
                }).
                when('/mypassword', {
                    templateUrl: 'views/mypassword.html'
                }).
                when('/myaddress', {
                    templateUrl: 'views/myaddress.html'
                }).
                when('/myreviewboard', {
                	templateUrl: 'views/myreviewboard.html'
                }).
                when('/myreviewboard/savedid/:SavedDataId', {
                    templateUrl: 'views/myreviewboard.html'
                }).
                when('/UsageTerm', {
                    templateUrl: 'views/usageterm.html'
                }).       
                when('/uploadfile/redirect/:ReUrl', {
                    templateUrl: 'views/dropboxupload.html'
                }).
                when('/dropboxupload', {
                    templateUrl: 'views/dropboxupload.html'
                }).
                when('/blogs', {
                    templateUrl: 'views/allblog.html'
                }).
                when('/blog/:NavName', {
                    templateUrl: 'views/viewblog.html'
                }).
                when('/writeblog', {
                    templateUrl: 'views/writeblog.html'
                }).
                when('/writeblog/savedid/:SavedDataId/uploaded/:UploadedUrls', {
                	templateUrl: 'views/writeblog.html'
                }).
                when('/page/:PageNum', {
                	templateUrl: 'views/page.html'
                }).
                when('/products', {
                    templateUrl: 'views/products.html'
                }).
                when('/addproduct', {
                    templateUrl: 'views/addproduct.html'
                }).
                when('/addproduct/savedid/:SavedDataId/uploaded/:UploadedUrls', {
                    templateUrl: 'views/addproduct.html'
                }).
                when('/product/:NavName', {
                    templateUrl: 'views/editproduct.html'
                }).
                when('/product/:NavName/savedid/:SavedDataId/uploaded/:UploadedUrls', {
                    templateUrl: 'views/editproduct.html'
                }).
                when('/addcustomer', {
                    templateUrl: 'views/addcustomer.html'
                }).
                when('/addcustomer/savedid/:SavedDataId/uploaded/:UploadedUrls', {
                    templateUrl: 'views/addcustomer.html'
                }).
                when('/customers', {
                    templateUrl: 'views/customers.html'
                }).
                when('/customer/:Id', {
                    templateUrl: 'views/editcustomer.html'
                }).        
                when('/customer/:Id/savedid/:SavedDataId/uploaded/:UploadedUrlsId', {
                    templateUrl: 'views/editcustomer.html'
                }).        
                when('/reports', {
                    templateUrl: 'views/reports.html'
                }).
                when('/enquires', {
                	templateUrl: 'views/enquires.html'
                }).
                when('/enquire/:Id', {
                	templateUrl: 'views/editenquire.html'
                }).
                when('/deal/:NavName', {
                    templateUrl: 'views/deal.html'
                }).
                when('/deals', {
                    templateUrl: 'views/deals.html'
                }).
                when('/', {
                    templateUrl: 'views/page.html'
                }).
                otherwise({
                    redirectTo: '/'
                });
    }
]);

//Setting HTML5 Location Mode
window.app.config(['$locationProvider',
    function($locationProvider) {
        $locationProvider.hashPrefix("!");
    }
]);
