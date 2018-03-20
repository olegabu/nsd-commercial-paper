/**
 * @class UserService
 * @classdesc
 * @ngInject
 */
function UserService($log, $rootScope, ApiService, localStorageService, ConfigLoader) {

  var config;

  var _refreshTokenTimer;

  ConfigLoader.ready().then(function(){

    config = ConfigLoader.get();

    if(UserService.getOrgRole() == '*'){
      $log.warn('Client role not set');
    } else {
      $log.info('Client role: ' + UserService.getOrgRole() );
    }

  });

  /**
   * @param {{username:string, orgName:string}} user
   */
  UserService.signUp = function(user) {
    $log.debug('UserService.signUp', user);
    return ApiService.user.signUp(user.username, user.orgName)
      .then(function(/** @type {TokenInfo} */tokenInfo){
        $log.debug('UserService.signUp result', tokenInfo);
        $rootScope._tokenInfo = tokenInfo; // used in http provider
        UserService.saveAuthorization(tokenInfo);
        return tokenInfo;
      });
  };

  UserService.refreshToken = function(tokenInfo){
    $log.debug('UserService.refreshToken', tokenInfo.tokenData);
    // TODO: now we use sign up method =(
    // return UserService.signUp({username: tokenData.username, orgName: tokenData.orgName});
    return UserService.signUp(tokenInfo.tokenData);
  }


  UserService.logout = function() {
    UserService.saveAuthorization(null);
  };


  UserService.isAuthorized = function(){
    return !!$rootScope._tokenInfo;
  }

  UserService.getUser = function(){
    return $rootScope._tokenInfo;
  }

  UserService.getOrgRole = function(){
    return config ? config['account-config'][config.org].role || '*' : null;
  }

  UserService.getOrg = function(){
    return config ? config.org : null;
  }

  /** 
   * @param {TokenInfo} user - actually tokenInfo
   */
  UserService.saveAuthorization = function(user){
    if(user){
      user.tokenData = parseTokenData(user.token) || {};
      UserService._setTokenExpTimeout(user);
    }
    localStorageService.set('user', user);
    $rootScope._tokenInfo = user;
  };

  UserService.restoreAuthorization = function(){
    var tokenInfo = localStorageService.get('user');

    if(tokenInfo){
      // {"exp":1500343472,"username":"test22","orgName":"org2","iat":1500307472}
      tokenInfo.tokenData = parseTokenData(tokenInfo.token) || {};

      if( (tokenInfo.tokenData.exp||0)*1000 <= Date.now() ){
        // token expired
        tokenInfo = null;
      }

      if(tokenInfo){
        UserService._setTokenExpTimeout(tokenInfo);
      }
    }

    $log.info('UserService.restoreAuthorization', !!tokenInfo);
    $rootScope._tokenInfo = tokenInfo;
  };


  /**
   *
   */
  UserService._setTokenExpTimeout = function(tokenInfo){
    var expiration = tokenInfo.tokenData.exp;
    $log.debug('UserService._setTokenExpTimeout', expiration);
    var timeout = expiration - Math.floor(Date.now() / 1000);
    if(timeout>0){
      timeout = Math.ceil(timeout * 0.9); // just a bit of time (10%) to update token
      $log.debug('UserService._setTokenExpTimeout set timeout for %s sec', timeout);

      if(_refreshTokenTimer){
        clearTimeout(_refreshTokenTimer);
      }
      _refreshTokenTimer = setTimeout(function(){
        _refreshTokenTimer = null;
        UserService.refreshToken(tokenInfo);
      }, timeout * 1000);
    } else {
      $log.warn('UserService._setTokenExpTimeout timeout has invalid value:', timeout);
    }
  }


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

  /**
   * Determine whether user can access the state.
   * Note, that it doesn't specified whether the user is authorizerd or not.
   * When user is not authorized, some states can be accessible.
   * @return {boolean}
   */
  UserService.canAccess = function(state){
    // check access
    var isAllowed = false;

    var rolesAllowed = state.data ? state.data.roles || '*' : '*';
    var role = UserService.getOrgRole();
    rolesAllowed = rolesAllowed || ['*'];
    if(rolesAllowed == '*' || rolesAllowed.indexOf(role) >= 0 ) {
      isAllowed = true;
    }
    // console.log('UserService.canAccess:', isAllowed, state.name);
    return isAllowed && UserService.isAuthorized();
  };

  return UserService;
}

/**
 * @ngModule
 */
angular.module('nsd.service.user', ['nsd.service.api','nsd.config.env', 'LocalStorageModule'])
  .service('UserService', UserService)

  .run(function(UserService, $log){
    UserService.restoreAuthorization();
  });
