/* globals angular */
/* jshint eqeqeq: false */
/* jshint -W014 */

/**
 * @param $scope
 * @param $q
 * @param $filter
 * @param {InstructionService} InstructionService
 * @param {BookService} BookService
 * @param {UserService} UserService
 * @param {DialogService} DialogService
 * @param {ConfigLoader} ConfigLoader
 * @constructor
 *
 * @class InstructionsController
 * @ngInject
 */
function InstructionsController($scope, $q, $filter, InstructionService, BookService, UserService, DialogService, ConfigLoader, Upload) {
  "use strict";

  var ctrl = this;
  ctrl.list = [];
  ctrl.redeemList = [];

  // var DATE_INPUT_FORMAT = 'dd/mm/yyyy';
  var DATE_FABRIC_FORMAT = 'yyyy-mm-dd'; // ISO
  var TRANSFER_SIDE_TRANSFERER = 'transferer';
  var TRANSFER_SIDE_RECEIVER = 'receiver';
  var NSD_ROLE = 'nsd';


  ctrl.org = ConfigLoader.get().org;
  ctrl.account = ConfigLoader.getAccount(ctrl.org);

  ctrl.uploadSignatureInstruction = null;
  /**
   * @type {boolean}
   */
  ctrl.test = ConfigLoader.get().dev;



  /**
   *
   */
  ctrl.init = function(){
      $scope.$on('chainblock', function(e, block) {
        if( InstructionService.isBilateralChannel(block.getChannel()) || block.getChannel() === BookService.getChannelID()){
          ctrl.reload();
        }
      });
      ctrl.reload();
  };

  /**
   *
   */
  ctrl.reload = function(){
    ctrl.invokeInProgress = true;
    return $q.all([
      InstructionService.listAll()
        .then(function(list){
          ctrl.list = list;
        }),

      UserService.getOrgRole() !== NSD_ROLE
      ? $q.resolve()
      : BookService.redeemHistory()
        .then(function(redeemList){
          ctrl.redeemList = redeemList;
        })
    ])
    .finally(function(){
      ctrl.invokeInProgress = false;
    });
  };


  ctrl.isTransferer = function(instruction){
    var acc = Object.keys(ctrl.account.acc);
    return acc.indexOf(instruction.transferer.account) > -1;
  };

  ctrl.isReceiver = function(instruction){
    var acc = Object.keys(ctrl.account.acc);
    return acc.indexOf(instruction.receiver.account) > -1;
  };

  /**
   * @param {Instruction} instruction
   * @param {boolean} getTheOppositeSide
   */
  ctrl.getInstructionID = function(instruction, getTheOppositeSide){
    if(instruction.type === 'dvp') {
      if (instruction.initiator === 'transferer' ^ getTheOppositeSide) {
        return 'INSTRUCTION_TRANSFER_DVP_ID';
      } else {
        return 'INSTRUCTION_RECEIVER_DVP_ID';
      }
    } else {
      // type === 'fop'
      if (instruction.initiator === 'transferer' ^ getTheOppositeSide) {
        return 'INSTRUCTION_TRANSFER_FOP_ID';
      } else {
        return  'INSTRUCTION_RECEIVER_FOP_ID';
      }
    }
  };


  ctrl.isInitiator = function(instruction){
    return instruction.initiator === 'transferer' ? ctrl.isTransferer(instruction) : ctrl.isReceiver(instruction);
  };


  ctrl.isSignAvailable = function(instruction, side) {
    var showSignStatuses = instruction.status=='executed'
      || instruction.status=='receiver-signed'
      || instruction.status=='transferer-signed'
      || instruction.status=='signed'
      || instruction.status=='downloaded';

    if (!showSignStatuses) { return false; }

    switch ( side ) {
      case 'transferer': return instruction.alamedaSignatureFrom;
      case 'receiver':   return instruction.alamedaSignatureTo;
      default: throw new Error('Unknown side: ' + side);
    }
  };



  ctrl.getSignLink = function(instruction, side) {
    var data;
    switch ( side ) {
      case 'transferer': data = instruction.alamedaSignatureFrom; break;
      case 'receiver':   data = instruction.alamedaSignatureTo; break;
      default: throw new Error('Unknown side: ' + side);
    }
    var base64data = data;
    var isBase64 = true;
    return 'data:application/octet-stream' + (isBase64 ? ';base64' : '' ) + ',' + base64data;
  };


  ctrl.getSignFilename = function(instruction, side) {
    // return instructionFilename(instruction, side) + '.xml';
    return instructionFilename(instruction, side) + '.sig';
  };


  ctrl.markSignDownloaded = function(instruction, side) {
    var defer = $q.resolve();
    if (side == 'transferer' && !instruction.transfererSignatureDownloaded) {
      defer = null;
    }
    if (side == 'receiver' && !instruction.receiverSignatureDownloaded) {
      defer = null;
    }

    if (!defer) {
      defer = $q.resolve()
        .then(function(){ ctrl.invokeInProgress = false; })
        .then(function(){
          return InstructionService.updateDownloadFlags(instruction, side);
        })
        .finally(function(){ ctrl.invokeInProgress = false; })
    }
  };


  ctrl.isAdmin = function(){
    return ctrl.org === NSD_ROLE;
  };



  ctrl.isInstructionXmlAvailable = function(instruction) {
    return instruction.status=='executed'
      || instruction.status=='receiver-signed'
      || instruction.status=='transferer-signed'
      || instruction.status=='signed'
      || instruction.status=='downloaded';
  };



  ctrl.getInstructionXmlLink = function(instruction, side, inverse) {
    var data;
    switch ( side ) {
      case 'transferer': data = inverse ? instruction.alamedaTo : instruction.alamedaFrom; break;
      case 'receiver':   data = inverse ? instruction.alamedaFrom : instruction.alamedaTo; break;
      default: throw new Error('Unknown side: ' + side);
    }
    var base64data = data;
    try {
      base64data = windows1251.encode(base64data);
    } catch(e) {
      console.warn('Fail to encode in windows-1251', base64data);
    }

    var isBase64 = true;
    try {
      base64data = btoa(base64data);
    } catch(e) {
      console.warn('Fail to encode in base64', base64data);
      isBase64 = false;
    }

    return 'data:application/xml' + (isBase64 ? ';base64' : '' ) + ',' + base64data;
    // return 'data:application/octet-stream;base64,' + base64data;
  };

  ctrl.oppositeSide = function(side) {
    if(side == 'transferer'){
       return 'receiver';
    } else if(side == 'receiver'){
       return 'transferer';
    } else {
      throw new Error('Unknown instruction side: ' + side);
    }
  };

  /**
   *
   */
  ctrl.getInstructionFilename = function(instruction, side) {
    return instructionFilename(instruction, side) + '.xml';
  };

  /**
   *
   */
  function instructionFilename(instruction, side) {

    var instructionCodeId = (side + '-' + instruction.type);

    var codes = {
      'transferer-fop'  : '16',
      'receiver-fop'    : '16.1',
      'transferer-dvp'  : '16.2',
      'receiver-dvp'    : '16.3'
    };

    var filenameTemplate = '%s-%s-%s';
    var args = [
      codes[instructionCodeId] || '00',
      instruction.reference,
      instruction.tradeDate.replace(/-/g, '')
    ];

    args.unshift(filenameTemplate);
    return format.apply(null, args);
  }

  /**
   *
   */
  function format(/*args*/) {
    var args = Array.prototype.slice.call(arguments);
    var str = args[0];
    for (var i = 1; i < args.length; i++) {
      str = str.replace('%s', args[i]);
    }
    return str;
  }

  /**
   *
   * @param inst instruction
   * @param type instruction type
   * @returns {boolean} true if instruction should be displayed
   */
  ctrl.showInstruction = function(inst, type) {
    var acc = Object.keys(ctrl.account.acc);
    return ctrl.org === NSD_ROLE ||
        (type === TRANSFER_SIDE_TRANSFERER && acc.indexOf(inst.transferer.account) > -1) ||
        (type === TRANSFER_SIDE_RECEIVER && acc.indexOf(inst.receiver.account) > -1);
        // (acc.indexOf(inst.transferer.account) > -1) || (acc.indexOf(inst.receiver.account) > -1);
  };

  /**
   * Displays reason based on role
   *
   * @param inst instruction
   * @param key object key that should be compared
   */
  ctrl.showReason = function(inst, key) {
    var curDep = inst[key];
    return ctrl.org === NSD_ROLE || curDep === ctrl.account.dep;
  };

  /**
   * @return {Instruction}
   */
  ctrl._getDefaultInstruction = function(transferSide, opponentID){
    var orgID = ctrl.org;
    return {
      transferer: {
        deponent: ctrl._getDeponentCode(transferSide === TRANSFER_SIDE_TRANSFERER ? orgID : opponentID)
      },
      receiver: {
        deponent: ctrl._getDeponentCode(transferSide === TRANSFER_SIDE_RECEIVER ? orgID : opponentID)
      },
      initiator: transferSide,
      // quantity: 0, // TODO: cause ui bug with overlapping label and input field with value
      tradeDate    : new Date(),
      instructionDate : new Date(),
      type: 'fop',


      transfererRequisites: {
        bic: '044525505'
      },
      receiverRequisites: {
        bic: '044525505'
      },
      paymentCurrency: 'RUB'

    };
  };

  ctrl._getDeponentCode = function(orgID){
    if(orgID === ctrl.org) {
      return ctrl.account.dep;
    }
    var account = ConfigLoader.getAccount(orgID) || {};
    return account.dep;
  };

  /**
   *
   */
  ctrl.getStatusClass = function(status){
    switch(status){
      case 'matched' : return 'deep-purple-text';
      case 'declined': return 'red-text darken-4';
      case 'executed': return 'green-text darken-4';
      case 'canceled': return 'grey-text';
      default: return '';
    }
  };

  /**
   * @param {Instruction} instruction
   * @return {boolean}
   */
  ctrl.canRollback = function(instruction) {
    return instruction.status === 'downloaded'
      || instruction.status === 'rollbackDeclined';
      // || instruction.status === 'executed'
      // || instruction.status === 'signed'

  };


  ctrl.canUploadSignature = function(instruction) {
    return instruction.status === 'executed'
       || instruction.status === 'transferer-signed'
       || instruction.status === 'receiver-signed'
       || instruction.status === 'signed';

  };

  /**
   * @param {Instruction} instruction
   */
  ctrl.rollbackInstruction = function(instruction){
    var cancelInstructionMessage = $filter('translate')('ROLLBACK_INSTRUCTION_PROMPT')
      .replace('%s', instruction.reference)
      .replace('%s', instruction.tradeDate);

    return DialogService.confirmReason(cancelInstructionMessage, {yesKlass:'red white-text'})
      .then(function(result){
        if(result.confirmed){
          ctrl.invokeInProgress = true;
          return InstructionService.rollbackInstruction(instruction, result.reason)
            .finally(function(){
              ctrl.invokeInProgress = false;
            });
        }
      });
  };

  ctrl.cancelInstruction = function(instruction){
    var cancelInstructionMessage = $filter('translate')('CANCEL_INSTRUCTION_PROMPT')
      .replace('%s', instruction.reference)
      .replace('%s', instruction.tradeDate);

    return DialogService.confirm(cancelInstructionMessage, {yesKlass:'red white-text'})
      .then(function(isConfirmed){
        if(isConfirmed){
          ctrl.invokeInProgress = true;
          return InstructionService.cancelInstruction(instruction)
            .finally(function(){
              ctrl.invokeInProgress = false;
            });
        }
      });

  };


  ctrl.confirmDownloaded = function(instruction) {
    ctrl.invokeInProgress = true;
    return InstructionService.setDownloaded(instruction)
            .finally(function(){
              ctrl.invokeInProgress = false;
            });
  };

  /**
   *
   */
  ctrl.newInstructionTransfer = function(transferSide, _channel){
    if(!$scope.inst || $scope.inst.initiator !== transferSide){
        // preset values

        var opponentOrgID = ctrl._getOrgIDByChannel(_channel);
        $scope.inst = ctrl._getDefaultInstruction(transferSide, opponentOrgID);
        $scope.formInstruction.$setPristine();
    }
  };


  ctrl.uploadSignatureDialog = function(instruction) {
    ctrl.uploadSignatureInstruction = instruction;

    var scope = {ctl:ctrl};
    return DialogService.dialog('upload-signature.html', scope);
  };


  ctrl.uploadSignature = function($file, cb){
    console.log('uploadSignature', $file);

    if (!$file) {
      return;
    }

    // ctrl._fileToString($file)
    ctrl.invokeInProgress = true;
    Upload.base64DataUrl($file)
      .then(function(base64uri){
        if (base64uri == 'data:') {
          // empty file was chosen
          throw new Error('File is empty');
        }
        return base64uri;
      })
      // remove mime type
      // leave pure base64 data
      .then(function(base64uri){ return base64uri.replace(/^.*base64,/, ''); })
      .then(function(base64data){
        console.log('file data:', base64data);
        cb();
        return InstructionService.sign(ctrl.uploadSignatureInstruction, base64data);
      })
      .finally(function(){
        ctrl.invokeInProgress = false;
      });
  };


  ctrl._fileToString = function(file) {
    return $q(function(resolve){
      var reader = new FileReader();
      reader.onload = function() {
          resolve(reader.result);
      };
      reader.readAsText(file);
    });
  };


  /**
   *
   */
  ctrl._getOrgIDByChannel = function(channelID){
    if(!channelID) {
      return null;
    }
    return channelID.split('-').filter(function(org){ return org !== ctrl.org; })[0];
  };

  /**
   *
   */
  ctrl.sendInstruction = function(instruction){
    $scope.inst = null;

    instruction.deponentFrom = instruction.transferer.deponent;
    instruction.deponentTo = instruction.receiver.deponent;

    // FIXME here date can come in two different formats:
    //  Date object when we change form value
    //  String (like '1 August, 2017') when we not change form value
    // Now we use formatDate() to transform both of it into ISO
    instruction.tradeDate        = formatDate(instruction.tradeDate);
    instruction.instructionDate  = formatDate(instruction.instructionDate);
    if(instruction.reason && instruction.reason.created){
      instruction.reason.created   = formatDate(instruction.reason.created);
    }

    var p;
    switch(instruction.initiator){
      case TRANSFER_SIDE_TRANSFERER:
        p = InstructionService.transfer(instruction);
        break;
      case TRANSFER_SIDE_RECEIVER:
        p = InstructionService.receive(instruction);
        break;
      default:
        throw new Error('Unknown transfer side: ' + instruction.initiator);
    }


    ctrl.invokeInProgress = true;
    return p.finally(function(){
      ctrl.invokeInProgress = false;
    });
  };

  /**
   * Parse date in format dd/mm/yyyy
   * @param {string|Date} date
   * @return {Date}
   */
  function formatDate(date) {
    if(!date) {
      return null;
    }

    if(!(date instanceof Date)){
      // assumind date is a string: '1 August, 2017'
      // TODO: we shouldn't rely on this
      date = new Date(date);
    }
    return date.format(DATE_FABRIC_FORMAT);
  }


  /**
   *
   */
  ctrl.newRedemption = function(){
    $scope.redemption = $scope.redemption || ctrl._getDefaultRedemption();
  };
  /**
   * @return {Redemption}
   */
  ctrl._getDefaultRedemption = function(){
    return {
      reason:{
        created   : new Date()//.format(DATE_INPUT_FORMAT)
      }
    };
  };

  /**
   * @param {Redemption} redemption
   */
  ctrl.sendRedemption = function(redemption){
    return DialogService.confirm( $filter('translate')('REDEEM_INSTRUCTION_PROMPT').replace('%s', redemption.security), {yesKlass:'red white-text'})
      .then(function(isConfirmed){
        if(isConfirmed){

          ctrl.invokeInProgress = true;
          return BookService.redeem(redemption)
            .finally(function(){
              ctrl.invokeInProgress = false;
            });
        }
      })
      .then(function(){
        $scope.redemption = null;
      });
  };


  /**
   * @param {Instruction} instruction
   */
  ctrl.showHistory = function(instruction){
    return InstructionService.history(instruction)
      .then(function(result){
        var scope = {history: result, getStatusClass: ctrl.getStatusClass, showReason: ctrl.showReason};
        return DialogService.dialog('balance-history.html', scope);
      });
  };


  // For prefill!
  ctrl.getOrgs = function(){
    var org = ConfigLoader.get().org;
    var accountConfig = ConfigLoader.get()['account-config'];
    var orgList = Object.keys(accountConfig)
      .filter(function(a){ return a!=='nsd'; })
      .filter(function(a){ return a!==org; })
      .sort(function(a, b){ return a.localeCompare(b); });
    return orgList;
  };



  ctrl.setPrefill = function(transferSide, from, to) {
    var org = ConfigLoader.get().org;

    var _prefillFrom = from || ctrl._prefillFrom || (transferSide == 'transferer' ? org : null);
    var _prefillTo = to || ctrl._prefillTo || (transferSide == 'receiver' ? org : null);

    if (_prefillFrom && _prefillTo) {
      $scope.inst = ctrl.getABStub(transferSide, _prefillFrom, _prefillTo)
    }
  };
  /**
   * @param transferSide
   * @return {Instruction}
   */
  ctrl.getABStub = function(transferSide, orgFrom, orgTo) {
    var accountConfig = ConfigLoader.get()['account-config'];

    var bicTransferer = Object.keys(accountConfig[orgFrom].bic)[0];
    var bicReceiver = Object.keys(accountConfig[orgTo].bic)[0];

    return {
      type:'dvp',

      transfererRequisites:{
        account: bicTransferer,
        bic: accountConfig[orgFrom].bic[bicTransferer] || '044525505'
      },
      receiverRequisites:{
        account: bicReceiver,
        bic: accountConfig[orgTo].bic[bicReceiver] || '044525505'
      },
      paymentAmount: 30000000,
      paymentCurrency: 'RUB',
      additionalInformation: transferSide === 'receiver' ? {description: 'payment no. DLT/001'} : null,

      security:'RU000A0JVVB5',
      transferer:{
        deponent: accountConfig[orgFrom].dep,
        account : Object.keys(accountConfig[orgFrom].acc)[0],
        division: accountConfig[orgFrom].acc[ Object.keys(accountConfig[orgFrom].acc)[0] ][0]
      },
      receiver:{
        deponent: accountConfig[orgTo].dep,
        account : Object.keys(accountConfig[orgTo].acc)[0],
        division: accountConfig[orgTo].acc[ Object.keys(accountConfig[orgTo].acc)[0] ][0]
      },
      initiator: transferSide,
      quantity: 1,
      reference: 'test',
      memberInstructionId: 123,
      tradeDate    : new Date(),
      instructionDate : new Date()
    };
  };


  //////////////

  // INIT
  ctrl.init();

}

angular.module('nsd.controller.instructions', ['nsd.service.instructions'])
.controller('InstructionsController', InstructionsController);

