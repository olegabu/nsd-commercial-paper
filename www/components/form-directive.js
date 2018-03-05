angular.module('nsd.directive.form', [])

  .directive('validateJson', function() {
    return {
      restrict:'A',
      require: 'ngModel',
      link: function(scope, elm, attrs, ctrl) {
        ctrl.$validators.json = function(modelValue, viewValue) {
          if (ctrl.$isEmpty(modelValue)) {
            return true;
          }

          try{
            JSON.parse(viewValue);
            // it is valid
            return true;
          }catch(e){
          // it is invalid
            return false;
          }

        };
      }
    };
  })

  .directive('script', function() {
    return {
      restrict:'E',
      scope: false,
      controller: function($scope, $attrs, $templateCache, $http, $log) {
        if($attrs['type'] != "text/ng-template" || !$attrs['src']){
          return;
        }

        var id = $attrs['id'] || $attrs['src'];
        var src = $attrs['src'];
        $log.debug('Loading %s template from %s', id, src);

        $http.get(src).then(function(response){
          $log.debug('Loaded template %s', id);
          $templateCache.put(id, response.data);
        });
      }
    };
  })

  .directive('ngLength', function() {
    return {
      require: 'ngModel',
      link: function(scope, elm, attrs, ctrl) {
        var length = parseInt(attrs['ngLength']);
        if(!length){
          throw new Error('ngLength should have a value' );
        }
        ctrl.$validators.length = function(modelValue, viewValue) {
          if (ctrl.$isEmpty(modelValue)) {
            // consider empty models to be valid
            return true;
          }

          return (''+viewValue).length == length;
        };
      }
    };
  })




  .directive('input', function() {
    return {
      restrict:'E',
      scope: false,
      require: '?ngModel',
      link: _inputLinkController
    };
  })
  .directive('textarea', function() {
    return {
      restrict:'E',
      scope: false,
      require: '?ngModel',
      link: _inputLinkController
    };
  });
  /**
   *
   */
  function _inputLinkController(scope, elm, attrs, ctrl) {

    // SET input id the same as name
    var inputName = elm.attr('name') || elm.attr('ng-model') || elm.attr('id');
    if(!elm.attr('name')){
      elm.attr('name', inputName);
    }
    if(!elm.attr('id')){
      elm.attr('id', inputName);
    }

    /*
    // SET ng-filled/ng-empty classes
    if (ctrl && ctrl.$validators && elm.attr('ng-model')) {
      // only apply the validator if ngModel is present
      scope.$watch( elm.attr('ng-model'), function(val){
        if( val ){
          elm.addClass('ng-filled').removeClass('ng-empty');
        } else {
          elm.removeClass('ng-filled').addClass('ng-empty');
        }
      });
    }
    */
  }