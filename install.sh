#!/bin/bash

#Package: obz
#Version: 1.0.0
#Section: utils
#Priority: optional
#Architecture: amd64
#Maintainer: Obaro I Johnson <johnson.obaro@hotmail.com>
#Description: Simple light tool for database migration

export PATH="$PATH:/usr/local/go/bin"
export PACKAGE_NAME=obz
export PACKAGE_VERSION="1.0.0"
export DIST_DIR="dist"
export BINARY_FILE_DIR="./bin"
export REPO_URI="https://github.com/miljimo/easymigration.git"
export HAS_GIT="$(type "git" &> /dev/null && echo true || echo false)"
export HAS_GO="$(type "go" &> /dev/null && echo true || echo false)"
export REPO_NAME="obz_code"



function  install(){


   rm -rf "$REPO_NAME"
   if [  "$HAS_GIT" != 'true'  ] ; then
     echo "required git to install this tools"
     return ;
   fi
   git clone  "$REPO_URI"  "$REPO_NAME"

   pushd "$REPO_NAME"
       echo "Repository = $REPO_NAME"
       echo "Current directory  = $(pwd)"
       # check if go is installed and if its installed it should run it.
       if [ $HAS_GO == 'true' ]; then
          ls -a .
          go mod tidy
          go build -o "$BINARY_FILE_DIR/obz"
       fi

       install_deb_package
   popd

   rm -rf "$REPO_NAME"
   echo "installation completed"
}

function install_deb_package(){
    local curr="${OSTYPE}"
    if [[ $curr == "linux-gnu"* ]]; then
      echo "supported"
      
       local parentDir="./$DIST_DIR/${PACKAGE_NAME}_${PACKAGE_VERSION}"
       sudo mkdir -p $parentDir
       sudo mkdir -p "$parentDir/usr/local/bin"
       sudo mkdir -p "$parentDir/DEBIAN"
       sudo mkdir -p "$DIST_DIR"

       # Create the control files and pupolate it with the information needed
       sudo touch  "$parentDir/DEBIAN/control"

       sudo chown -R $(whoami):$(whoami) dist
       echo "Package: ${PACKAGE_NAME}" >  "$parentDir/DEBIAN/control"
       echo "Version: ${PACKAGE_VERSION}" >>  "$parentDir/DEBIAN/control"
       echo "Section: utils" >>  "$parentDir/DEBIAN/control"
       echo "Architecture: amd64" >>  "$parentDir/DEBIAN/control"       
       echo "Priority: optional" >>  "$parentDir/DEBIAN/control"
       echo "Maintainer: Obaro I Johnson <johnson.obaro@hotmail.com>" >>  "$parentDir/DEBIAN/control"
       echo "Description: Simple light tool for database migration" >> "$parentDir/DEBIAN/control"  

       # At this point the application is already build and we just want to copy all the
       # binaries to the  usr/local/bin directories

       sudo cp -r ${BINARY_FILE_DIR}/* "$parentDir/usr/local/bin/"

       # now we have build the application we need to build the package
       local packageFile="./${DIST_DIR}/${PACKAGE_NAME}_${PACKAGE_VERSION}.deb"
       sudo dpkg-deb --build $parentDir  "$packageFile"
       #rm  -rf "$parentDir"


       #install the application into the linux environment 

       dpkg --install  $packageFile;
       echo "installation completed"
       return "$?";
    fi

    return 0
}

install

