angular.module("featen.share").controller("DropboxUploadController", ["$scope", "$rootScope", "$route", "$routeParams", "$location", "Global", "StageData", "Shares", "Alerts", function($scope, $rootScope, $route, $routeParams, $location, Global, StageData, Shares, Alerts) {
        var redirectUrl = $routeParams.ReUrl;

        var dropbox = document.getElementById("dropbox");
        $scope.dropText = '将文件拉拽到这个地方...';

        // init event handlers
        function dragEnterLeave(evt) {
            evt.stopPropagation();
            evt.preventDefault();
            $scope.$apply(function() {
                $scope.dropText = '将文件拉拽到这个地方...';
                $scope.dropClass = '';
            });
        }
        dropbox.addEventListener("dragenter", dragEnterLeave, false);
        dropbox.addEventListener("dragleave", dragEnterLeave, false);
        dropbox.addEventListener("dragover", function(evt) {
            evt.stopPropagation();
            evt.preventDefault();
            var clazz = 'not-available';
            var ok = evt.dataTransfer && evt.dataTransfer.types && evt.dataTransfer.types.indexOf('Files') >= 0;
            $scope.$apply(function() {
                $scope.dropText = ok ? '将文件拉拽到这个地方...' : '只支持文件';
                $scope.dropClass = ok ? 'over' : 'not-available';
            });
        }, false);
        dropbox.addEventListener("drop", function(evt) {
            console.log('drop evt:', JSON.parse(JSON.stringify(evt.dataTransfer)));
            evt.stopPropagation();
            evt.preventDefault();
            $scope.$apply(function() {
                $scope.dropText = '将文件拉拽到这个地方...';
                $scope.dropClass = '';
            });
            var files = evt.dataTransfer.files;
            if (files.length > 0) {
                $scope.$apply(function() {
                    $scope.files = [];
                    for (var i = 0; i < files.length; i++) {
                        $scope.files.push(files[i]);
                    }
                });
            }
        }, false);
        //============== DRAG & DROP =============

        $scope.setFiles = function(element) {
            $scope.$apply(function($scope) {
                $scope.files = [];
                for (var i = 0; i < element.files.length; i++) {
                    $scope.files.push(element.files[i]);
                }
                $scope.progressVisible = false;
            });
        };

        $scope.uploadFile = function() {
            var fd = new FormData();
            for (var i in $scope.files) {
                fd.append("files", $scope.files[i]);
            }
            var xhr = new XMLHttpRequest();
            xhr.upload.addEventListener("progress", uploadProgress, false);
            xhr.addEventListener("load", uploadComplete, false);
            xhr.addEventListener("error", uploadFailed, false);
            xhr.addEventListener("abort", uploadCanceled, false);
            xhr.open("POST", "/service/uploadphoto");
            $scope.progressVisible = true;
            xhr.send(fd);
        };

        function uploadProgress(evt) {
            $scope.$apply(function() {
                if (evt.lengthComputable) {
                    $scope.progress = Math.round(evt.loaded * 100 / evt.total);
                } else {
                    $scope.progress = 'Full speed uploading...';
                }
            });
        }

        function redirect(uploadedfileurls) {
            var enurl = redirectUrl;
            var deurl = Base64.decode(enurl);
            if (uploadedfileurls !== '') {
                var id = StageData.add(uploadedfileurls);
                $location.path(deurl + "/uploaded/" + id);
            } else {
                $location.path(deurl );
            }
            $rootScope.$apply();
        }

        function uploadComplete(evt) {
            Alerts.add("success", "Success!", "Upload success, return to previous page.");
            redirect(evt.currentTarget.responseText);
        }

        function uploadFailed(evt) {
            Alerts.add("error", "Failed!", "Upload failed, please check file size and network.");
            redirect('');
        }

        function uploadCanceled(evt) {
            $scope.$apply(function() {
                $scope.progressVisible = false;
            });
            Alerts.add("error", "Canceled!", "上传取消...");
        }
    }]);
