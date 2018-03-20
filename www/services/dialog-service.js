/* globals angular,$ */
/* jshint -W014 */

/**
 * @class DialogService
 * @classdesc
 * @ngInject
 */
function DialogService($document, $compile, $templateCache, $rootScope, $q, $log) {
  "use strict";

  // jshint shadow: true
  var DialogService = this;

  var confirmForm =
    '<div id="confirmDialog" class="modal modal-fixed-footer modal-dialog">'
    + '<form name="form" class="form-horizontal" novalidate>'
    + '<div class="modal-content">'
    + '<h4 translate>{{$options.title}}</h4>'
    + '<div class="row">{{$options.text}}</div>'
    + '</div>'
    + '<div class="modal-footer">'
    + '<button type="submit" class="modal-action waves-effect waves-green btn-flat {{$options.yesKlass}}" ng-click="$close(true)" translate>{{$options.yesLabel}}</button>'
    + '<button type="button" class="modal-action waves-effect waves-green btn-flat {{$options.noKlass}}" ng-click="$close(false)" translate>{{$options.noLabel}}</button>'
    + '</div>'
    + '</form>'
    + '</div>';

  $templateCache.put('confirmDialog.html', confirmForm);


  var confirmReasonForm =
    '<div id="confirmReasonDialog" class="modal modal-fixed-footer modal-dialog modal-prompt">'
    + '<form name="form" class="form-horizontal" novalidate>'
    + '<div class="modal-content">'
    + '<h4 translate>{{$options.title}}</h4>'
    + '<div class="row">{{$options.text}}</div>'
    + '<div class="row">'
    + '<div class="input-field col s12">'
    + '<textarea id="d-cr-reason" name="d-cr-reason" class="materialize-textarea" ng-model="$options.reason" required></textarea>'
    + '<label for="d-cr-reason" translate>{{$options.promptLabel}}</label>'
    + '</div>'
    + '</div>'
    + '</div>'
    + '<div class="modal-footer">'
    + '<button type="submit" class="modal-action waves-effect waves-green btn-flat {{$options.yesKlass}}" ng-disabled="form.$invalid" ng-click="$close({confirmed:true, reason:$options.reason})" translate>{{$options.yesLabel}}</button>'
    + '<button type="button" class="modal-action waves-effect waves-green btn-flat {{$options.noKlass}}" ng-click="form.$setPristine(); $close({confirmed:false, reason:$options.reason})" translate>{{$options.noLabel}}</button>'
    + '</div>'
    + '</form>'
    + '</div>';

  $templateCache.put('confirmReasonDialog.html', confirmReasonForm);

  // var element = $(confirmForm).appendTo($document[0].body);
  // var compiledConfirmForm = $compile(element.contents());

  /**
   * Create confirmation dialog
   * @param {string} text
   * @param [options]
   * @return {Promise<boolean>} resolves with true value when user has confirmed an action
   */
  DialogService.confirm = function(text, options){
    options = options || {};

    options.title = options.title || 'dialog.CONFIRM_TITLE' || 'Confirm your actions';
    options.text = text || options.text || 'Are you sure?';
    options.yesLabel = options.yesLabel || 'dialog.CONFIRM_ACTION' || 'Yes';
    options.yesKlass = options.yesKlass || '';
    options.noLabel  = options.noLabel || 'dialog.DECLINE_ACTION' || 'No';
    options.noKlass  = options.noKlass || '';

    return DialogService.dialog('confirmDialog.html', options);
    // .then(function(isConfirmed){
    //   return isConfirmed;
    // })
  };


  /**
   * Create confirmation dialog
   * @param {string} text
   * @param [options]
   * @return {Promise<boolean>} resolves with true value when user has confirmed an action
   */
  DialogService.confirmReason = function(text, options){
    options = options || {};

    options.title = options.title || 'dialog.PROMPT_TITLE' || 'Confirm your actions';
    options.text = text || options.text || 'Are you sure?';
    options.promptLabel = options.promptLabel || 'dialog.PROMPT_LABEL' || 'Enter reason here';
    options.yesLabel = options.yesLabel || 'dialog.CONFIRM_ACTION' || 'Yes';
    options.yesKlass = options.yesKlass || '';
    options.noLabel  = options.noLabel || 'dialog.DECLINE_ACTION' || 'No';
    options.noKlass  = options.noKlass || '';

    return DialogService.dialog('confirmReasonDialog.html', options);
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
   * @param {object} [options]
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
    var scope = $rootScope.$new(false);
    scope.$options = options;

    // TODO: optimize compilation
    var compiledElement = $compile(element.contents());
    compiledElement(scope);

    return $q(function(resolve){
      scope.$close = resolve;
      // element.leanModal(options); // ?
      element.openModal();
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
