/**
 * @class UserService
 * @classdesc
 * @ngInject
 */
function UserService($log, $rootScope, ApiService, localStorageService, env) {

  /**
   * @param {{username:string, orgName:string}} user
   */
  UserService.signUp = function(user) {
    return ApiService.user.signUp(user.username, user.orgName)
      .then(function(/** @type {TokenInfo} */tokenInfo){
        $rootScope._tokenInfo = tokenInfo; // used in http provider
        UserService.saveAuthorization(tokenInfo);
        return tokenInfo;
      });
  };


  UserService.logout = function() {
    UserService.saveAuthorization(null);
  };


  UserService.isAuthorized = function(){
    return !!$rootScope._tokenInfo;
  }

  UserService.getUser = function(){
    return $rootScope._tokenInfo;
  }


  UserService.saveAuthorization = function(user){
    if(user){
      user.tokenData = parseTokenData(user.token);
    }
    localStorageService.set('user', user);
  };

  UserService.restoreAuthorization = function(){
    var tokenInfo = localStorageService.get('user');
    $log.info('UserService.restoreAuthorization', !!tokenInfo);

    if(tokenInfo){
      // {"exp":1500343472,"username":"test22","orgName":"org2","iat":1500307472}
      tokenInfo.tokenData = parseTokenData(tokenInfo.token);
      // TODO: check expire time

      if( (tokenInfo.tokenData.exp||0)*1000 <= Date.now() ){
        // token expired
        tokenInfo = null;
      }
    }
    $rootScope._tokenInfo = tokenInfo;
  };


  /**
   *
   */
  function parseTokenData(token){
      token = token || "";
      var tokenDataEncoded = token.split('.')[1];
      var tokenData = null;
      try{
        tokenData = JSON.parse(atob(tokenDataEncoded));
      }catch(e){
        $log.warn(e);
      }
      return tokenData;
  }


  UserService.canAccess = function(state){
    // check access
    var isAllowed = false;

    var rolesAllowed = state.data ? state.data.roles || '*' : '*';
    rolesAllowed = rolesAllowed || ['*'];
    if(rolesAllowed == '*' || rolesAllowed.indexOf(env.role) >= 0 ) {
      isAllowed = true;
    }
    // console.log('UserService.canAccess:', isAllowed, state.name);
    return isAllowed;
  };

  return UserService;
}

angular.module('nsd.service.user', ['nsd.service.api','nsd.config.env', 'LocalStorageModule'])
  .service('UserService', UserService)

  .run(function(UserService, $log, env){
    UserService.restoreAuthorization();

    if(!env.role){
      $log.warn('Client role not set');
      env.role = '*';
    } else {
      $log.info('Client role:' + env.role);
    }
  });
