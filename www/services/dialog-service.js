/**
 * @class DialogService
 * @classdesc
 * @ngInject
 */
function DialogService($document, $compile, $rootScope, $q, $log) {

  // jshint shadow: true
  var DialogService = this;

  var confirmForm =
    '<div id="confirmDialog" class="modal modal-fixed-footer modal-dialog">'
    + '<form name="form" class="form-horizontal" novalidate>'
      + '<div class="modal-content">'
        + '<h4>{{title}}</h4>'
        + '<div class="row">{{text}}</div>'
      + '</div>'
      + '<div class="modal-footer">'
        + '<input type="submit" class="modal-action modal-close waves-effect waves-green btn-flat red white-text" ng-click="confirm()" value="{{confirmBtn}}"/>'
        + '<input type="submit" class="modal-action modal-close waves-effect waves-red btn-flat" ng-click="reject()" value="No"/>'
      + '</div>'
    + '</form>'
  + '</div>';


  var element = $(confirmForm).appendTo($document[0].body);
  var compiledConfirmForm = $compile(element.contents());


  DialogService.confirm = function(text, options){
    options = options || {};

    // $compile(element.contents())(scope);
    var scope = $rootScope.$new();
    scope.title = 'Confirm your actions';
    scope.text = text || 'Are you sure?';
    scope.confirmBtn = options.confirmBtn || 'Yes';


    return $q(function(resolve, reject){
      scope.confirm = resolve;
      scope.reject = reject;

      compiledConfirmForm(scope);

      var complete = function () {
          scope.open = false;
          scope.$apply();
      };
      var options = {
          complete: complete,
      };
      // element.leanModal(options);
      element.openModal(options)
    });
  }



  return DialogService;
}

angular.module('nsd.service.dialog', [])
  .service('DialogService', DialogService);
