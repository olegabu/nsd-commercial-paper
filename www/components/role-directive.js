(function(){

  angular.module('nsd.directive.role', ['nsd.service.user'])

    /**
     * @ngDirective role-show

     * @example <div role-show="sales">...</div>
     * @example <div role-show="sales, investor">...</div>
     * @example <div role-show="*">...</div>
     */
    .directive('roleShow', function(UserService) {
      return {
        restrict:'A',
        link: function(scope, elm, attrs, ctrl) {
          // console.log('roleShow', attrs);
          var roles = (attrs['roleShow']||"");
          var orgRole = UserService.getOrgRole();
          if(!_matchRole(orgRole, roles)){
            elm.hide();
          }
        }
      };
    })

    .directive('roleHide', function(UserService) {
      return {
        restrict:'A',
        link: function(scope, elm, attrs, ctrl) {
          // console.log('roleHide', attrs);
          var roles = (attrs['roleHide']||"");
          var orgRole = UserService.getOrgRole();
          if(_matchRole(orgRole, roles)){
            elm.hide();
          }
        }
      };
    });


    function _matchRole(role, rolesAttributeValue){
        // console.log('_matchRole', role, 'against', rolesAttributeValue);
        if(rolesAttributeValue.trim() == '*'){
          return true;
        }

        var roles = rolesAttributeValue.split(',').map(function(role){ return role.trim(); });
        return roles.indexOf(role) >= 0;
    }

})();
