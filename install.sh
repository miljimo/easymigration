#!/bin/bash

#Package: obz
#Version: 1.0.0
#Section: utils
#Priority: optional
#Architecture: amd64
#Maintainer: Obaro I Johnson <johnson.obaro@hotmail.com>
#Description: Simple light tool for database migration

export PACKAGE_NAME=obz
export PACKAGE_VERSION="1.0.0"
export DIST_DIR="dist"
export BINARY_FILE_DIR="./bin"


function install_deb_package(){
    local curr="${OSTYPE}"
    if [[ $curr == "linux-gnu"* ]]; then
      echo "supported"
      
       local parentDir="./$DIST_DIR/${PACKAGE_NAME}_${PACKAGE_VERSION}"
       mkdir -p $parentDir
       mkdir -p "$parentDir/usr/local/bin"
       mkdir -p "$parentDir/DEBIAN"
       mkdir -p "$DIST_DIR"

       # Create the control files and pupolate it with the information needed
       touch  "$parentDir/DEBIAN/control"
       echo "Package: ${PACKAGE_NAME}" >  "$parentDir/DEBIAN/control"
       echo "Version: ${PACKAGE_VERSION}" >>  "$parentDir/DEBIAN/control"
       echo "Section: utils" >>  "$parentDir/DEBIAN/control"
       echo "Architecture: amd64" >>  "$parentDir/DEBIAN/control"       
       echo "Priority: optional" >>  "$parentDir/DEBIAN/control"
       echo "Maintainer: Obaro I Johnson <johnson.obaro@hotmail.com>" >>  "$parentDir/DEBIAN/control"
       echo "Description: Simple light tool for database migration" >> "$parentDir/DEBIAN/control"  

       # At this point the application is already build and we just want to copy all the
       # binaries to the  usr/local/bin directories

       cp -r ${BINARY_FILE_DIR}/* "$parentDir/usr/local/bin/"

       # now we have build the application we need to build the package
       local packageFile="./${DIST_DIR}/${PACKAGE_NAME}_${PACKAGE_VERSION}.deb"
       dpkg-deb --build $parentDir  "$packageFile"
       #rm  -rf "$parentDir"


       #install the application into the linux environment 

       dpkg --install  $packageFile;
       echo "installation completed"
       return "$?";
    fi

    return 0
}


install_deb_package
