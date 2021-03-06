Integration with the EGI Federated Cloud
========================================

`EGI <https://www.egi.eu/>`_ is a federation of many cloud providers and hundreds of data centres, spread across Europe and worldwide that delivers advanced computing services to support scientists, multinational projects and research infrastructures.

The `EGI Federated Cloud <https://www.egi.eu/federation/egi-federated-cloud/>`_ is an IaaS-type cloud, made of academic private clouds and virtualised resources and built around open standards. Its development is driven by requirements of the scientific communities.

EGI Applications on Demand: EC3 Portal
--------------------------------------

The OSCAR platform can be deployed on the EGI Federated Cloud resources through the `EC3 Portal <https://servproject.i3m.upv.es/ec3-ltos/index.php>`_ available in the `EGI Applications on Demand <https://www.egi.eu/services/applications-on-demand/>`_ service.

The `EC3 Web Interface documentation <https://ec3.readthedocs.io/en/devel/ec3aas.html>`_ can be followed in order to deploy the platform. Remember to pick “OSCAR” as the Local Resource Management System (LRMS).

.. image:: images/oscar-egi-ec3.png
   :scale: 60 %

EGI DataHub
-----------

`EGI DataHub <https://datahub.egi.eu/>`_, based on `Onedata <https://onedata.org/#/home>`_, provides a global data access solution for science. Integrated with the EGI AAI, it allows users to have Onedata spaces supported by providers across Europe for replicated storage and on-demand caching. 

EGI DataHub can be used as a storage provider and source of events for OSCAR, allowing users to process their files by uploading them to a Onedata space. This can be done thanks to the development of:

-  `OneTrigger <https://github.com/grycap/onetrigger>`_. A command-line tool to detect Onedata file events in order to trigger a webhook (i.e. an OSCAR Function).
-  `FaaS-Supervisor <https://github.com/grycap/faas-supervisor>`_. Used in OSCAR and `SCAR <https://github.com/grycap/scar>`_, responsible for managing the data Input/Output and the user code execution. Support for Onedata has been added to perform the integration with EGI DataHub.

.. image:: images/oscar-onetrigger.png
   :scale: 60 %

To deploy a function with Onedata support you only have to specify the URL of the Oneprovider host, your access token and the space name where files will be stored.
This will trigger the following:

-  Creation of input and output folders in the specified space.
-  Deployment of a OneTrigger pod for the function.
-  Injection of Onedata’s login variables into the function.

.. image:: images/oscar-onedata.png
   :scale: 60 %


This means that scientists can upload input files to their Onedata space in the EGI DataHub in order to automatically trigger the execution of a function in order to process this file.
Multiple file uploads results in multiple function invocations that run as Kubernetes job in the elastic Kubernetes cluster dynamically provisioned from the EGI Federated Cloud.

Video Demo
-----------
This video demo will get you up & running deploying an OSCAR cluster, through the EGI Applications on Demand portal, in the EGI Federated Cloud and using EGI DataHub as the source of events to perform the automated parallel classification of `plants using deep learning techniques <https://github.com/deephdc/DEEP-OC-plant-classification-theano>`_, a use case from the `DEEP Hybrid-DataCloud <http://www.deep-hybrid-datacloud.eu>`_ project.
You will only need a valid user account in EGI to follow the steps.

.. raw:: html

    <iframe width="560" height="315" src="https://www.youtube.com/embed/ZtAlVc1uLwc" frameborder="0" allowfullscreen></iframe>