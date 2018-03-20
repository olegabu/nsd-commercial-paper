/* globals angular */

angular.module('nsd.directive.payment', [])

  .directive('instructionPayment', function(ConfigLoader) {
    "use strict";


    return {
      require: ['^^form'],
      restrict:'E',
      scope: {
        label:'@',
        form:"=",
        requsite:'=model',
        deponent:"="
      },
      templateUrl:'components/instruction/payment.html',
      link: function(scope, elm, attrs, ctrl) {
        scope.form = ctrl[0]; // form
        scope.componentId = scope.$id;

        var accountData = ConfigLoader.getAccounts();

        scope.accountData = accountData;
        scope.allAccounts = Object.keys(accountData).reduce(function(result, orgID) {
          var deponent = accountData[orgID].dep;
          result[deponent] = accountData[orgID];
          return result;
        }, {});

      }
    };
  });