/**
 * @class DialogService
 * @classdesc
 * @ngInject
 */
function DialogService($document, $compile, $templateCache, $rootScope, $q, $log) {

  // jshint shadow: true
  var DialogService = this;

  var confirmForm =
    '<div id="confirmDialog" class="modal modal-fixed-footer modal-dialog">'
    + '<form name="form" class="form-horizontal" novalidate>'
      + '<div class="modal-content">'
        + '<h4>{{$options.title}}</h4>'
        + '<div class="row">{{$options.text}}</div>'
      + '</div>'
      + '<div class="modal-footer">'
        + '<input type="submit" class="modal-action waves-effect waves-green btn-flat {{$options.yesKlass}}" ng-click="$close(true)" value="{{$options.yesLabel}}"/>'
        + '<input type="button" class="modal-action waves-effect waves-green btn-flat {{$options.noKlass}}" ng-click="$close(false)" value="{{$options.noLabel}}"/>'
      + '</div>'
    + '</form>'
  + '</div>';

  $templateCache.put('confirmDialog.html', confirmForm);

  // var element = $(confirmForm).appendTo($document[0].body);
  // var compiledConfirmForm = $compile(element.contents());


  DialogService.confirm = function(text, options){
    options = options || {};

    options.title = options.title  || 'Confirm your actions';
    options.text = text || options.text || 'Are you sure?';
    options.yesLabel = options.yesLabel || 'Yes';
    options.yesKlass = options.yesKlass || '';
    options.noLabel  = options.noLabel || 'No';
    options.noKlass  = options.noKlass || '';

    return DialogService.dialog('confirmDialog.html', options);
      // .then(function(isConfirmed){
      //   return isConfirmed;
      // })
  };


  // DIALOG SERVICE IS INCOMPLETE!

  /**
   * Create an arbitary dialog
   *  Dialog has $options and $close
   *
   * @param {string} dialogID
   * @param {Scope} [scope]
   * @return {Promise}
   */
  DialogService.dialog = function(dialogID, options){

    // $compile(element.contents())(scope);
    var template = $templateCache.get(dialogID);
    if(!template){
      throw new Error('Template not defined: '+dialogID);
    }
    var element = $(template).appendTo($document[0].body);

    // create a scope
    var scope = $rootScope.$new();
    scope.$options = options;

    // TODO: optimize compilation
    var compiledElement = $compile(element.contents());
    compiledElement(scope);

    return $q(function(resolve, reject){
      scope.$close = resolve;
      // element.leanModal(options); // ?
      element.openModal()
    })
    .finally(function(){
      element.closeModal();
      scope.$destroy();
    });
  };



  return DialogService;
}

angular.module('nsd.service.dialog', [])
  .service('DialogService', DialogService);
