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


  .directive('input', function() {
    return {
      restrict:'E',
      require: '?ngModel',
      link: function(scope, elm, attrs, ctrl) {
        var inputName = elm.attr('name') || elm.attr('ng-model');
        if(!elm.attr('name')){
          elm.attr('name', inputName);
        }
        if(!elm.attr('id')){
          elm.attr('id', inputName);
        }
        // TODO: form name hardcoded

        // scope.$watch('form.'+inputName+'.$invalid && form.'+inputName+'.$dirty', function(val){
        //   if(val){
        //     elm.addClass('invalid');
        //   } else {
        //     elm.removeClass('invalid');
        //   }
        // });


        // // only apply the validator if ngModel is present and AngularJS has added the email validator
        // if (ctrl && ctrl.$validators) {

        //   scope.

        // }
      }
    };
  });